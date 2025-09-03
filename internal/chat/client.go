package chat

import (
    "encoding/json"
    "log"
    "time"
    "net/http"
    "strconv"

    "github.com/gorilla/websocket"
    "github.com/kaelCoding/toyBE/internal/database"
    "github.com/kaelCoding/toyBE/internal/models"
    "github.com/kaelCoding/toyBE/internal/fcm"
    "github.com/gin-gonic/gin"
)

const (
    writeWait = 10 * time.Second
    pongWait = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
    maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{"http://localhost:5173", "https://tunitoku.netlify.app", "https://tunitoku.store"}
        for _, o := range allowedOrigins {
            if o == origin {
                return true
            }
        }
        return false
    },
}

type Client struct {
    hub *Hub
    conn *websocket.Conn

    send chan []byte
    userID uint
    isAdmin bool
}

type ChatMessage struct {
    ReceiverID uint   `json:"receiverId"`
    Content    string `json:"content"`
}

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }

        var msg ChatMessage
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("Error unmarshalling message: %v", err)
            continue
        }

        db := database.GetDB()
        var adminUser models.User
        db.Where("admin = ?", true).First(&adminUser)
        adminID := adminUser.ID

        var receiverID uint
        if c.isAdmin {
            receiverID = msg.ReceiverID
        } else {
            receiverID = adminID
        }

        dbMessage := models.Message{
            SenderID:   c.userID,
            ReceiverID: receiverID,
            Content:    msg.Content,
            Timestamp:  time.Now(),
        }

        if err := db.Create(&dbMessage).Error; err != nil {
            log.Printf("Error saving message to DB: %v", err)
            continue
        }

        var senderUser models.User
        db.First(&senderUser, c.userID)

        recipient, ok := c.hub.clients[receiverID]
        if ok {
            fullMessage, _ := json.Marshal(dbMessage)
            select {
            case recipient.send <- fullMessage:
            default:
                close(recipient.send)
                delete(c.hub.clients, recipient.userID)
            }
        } else {
            var recipientUser models.User
            db.First(&recipientUser, receiverID)
            
            if recipientUser.FCMToken != "" {
                fcm.SendNotification(
                    recipientUser.FCMToken,
                    "Tin nhắn mới từ " + senderUser.Username,
                    msg.Content,
                    strconv.FormatUint(uint64(c.userID), 10),
                )
            }
        }
        
        fullMessage, _ := json.Marshal(dbMessage)
        c.send <- fullMessage
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write(<-c.send)
            }

            if err := w.Close(); err != nil {
                return
            }
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func ServeWs(hub *Hub, c *gin.Context, userID uint, isAdmin bool) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Println(err)
        return
    }
    client := &Client{
        hub:     hub,
        conn:    conn,
        send:    make(chan []byte, 256),
        userID:  userID,
        isAdmin: isAdmin,
    }
    client.hub.register <- client

    go client.writePump()
    go client.readPump()
}
