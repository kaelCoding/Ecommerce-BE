package fcm

import (
	"context"
	"log"
	"time"
	"os"

	"firebase.google.com/go/v4/messaging"
	"firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var client *messaging.Client

func InitializeFirebaseClient() {
	ctx := context.Background()

	serviceAccountKeyJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if serviceAccountKeyJSON == "" {
		log.Fatalf("Lỗi: Biến môi trường 'FIREBASE_SERVICE_ACCOUNT_KEY' không được đặt.")
	}

	sa := option.WithCredentialsJSON([]byte(serviceAccountKeyJSON))
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err = app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v", err)
	}
	log.Println("Firebase Messaging client initialized successfully.")
}

func SendNotification(fcmToken, title, body, senderID string) {
	if client == nil {
		log.Println("FCM client is not initialized.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: map[string]string{
			"sender_id": senderID,
		},
		Token: fcmToken,
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}
	log.Printf("Successfully sent message: %s", response)
}

