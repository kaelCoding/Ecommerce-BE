package handlers

import (
	"bytes"
	"crypto/ecdsa"    
	"crypto/elliptic" 
	"crypto/rand"     
	"encoding/base64" 
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync" 
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5" 
	"github.com/google/uuid"
)

type SearchItem struct {
    Name      string  `json:"name"`
    PriceJPY  float64 `json:"priceJPY"`
    ImageURL  string  `json:"imageURL"`
    ItemURL   string  `json:"itemURL"`
    IsSoldOut bool    `json:"isSoldOut"`
}
type SearchRequest struct {
    Keyword string `json:"keyword" binding:"required"`
}
type FetchRequest struct {
    URL string `json:"url" binding:"required"`
}
type FetchResponse struct {
    Name          string   `json:"name"`
    PriceJPY      float64  `json:"priceJPY"`
    Description   string   `json:"description"`
    Condition     string   `json:"condition"`
    ImageURLs     []string `json:"imageURLs"`
    MercariItemID string   `json:"mercariItemID"`
}
type MercariSearchCondition struct {
    Keyword        string   `json:"keyword"`
    ExcludeKeyword string   `json:"excludeKeyword"`
    Sort           string   `json:"sort"`
    Order          string   `json:"order"`
    Status         []string `json:"status"`
}
type MercariSearchPayload struct {
    PageSize        int                    `json:"pageSize"`
    PageToken       string                 `json:"pageToken"`
    SearchSessionId string                 `json:"searchSessionId"`
    Source          string                 `json:"source"`
    SearchCondition MercariSearchCondition `json:"searchCondition"`
    WithItemBrand   bool                   `json:"withItemBrand"`
}
type MercariAPIItem struct {
    ID         string   `json:"id"`
    Name       string   `json:"name"`
    Price      string   `json:"price"` 
    Thumbnails []string `json:"thumbnails"`
    Status     string   `json:"status"`
}
type MercariAPIResponse struct {
    Items []MercariAPIItem `json:"items"`
    Meta  struct {
        NumFound string `json:"numFound"`
    } `json:"meta"`
}
type MercariItemAttribute struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}
type MercariItemDetail struct {
    ID             string                 `json:"id"`
    Name           string                 `json:"name"`
    Price          float64                `json:"price"` 
    Description    string                 `json:"description"`
    Status         string                 `json:"status"`
    Photos         []string               `json:"photos"` 
    ItemAttributes []MercariItemAttribute `json:"itemAttributes"`
}
type MercariItemResponse struct {
    Data MercariItemDetail `json:"data"`
}

var (
	priceRegex  = regexp.MustCompile(`[¥,() .]`)
	itemIDRegex = regexp.MustCompile(`(m\d+)`)
	httpClient  = &http.Client{
		Timeout: 30 * time.Second,
	}

	mercariPrivateKey *ecdsa.PrivateKey
	mercariJwk        json.RawMessage 
	mercariUUID = "313daa9c-c2a8-49ea-8584-742c0286c512" 
	keyGenOnce  sync.Once
)

func initKeys() {
	keyGenOnce.Do(func() {
		var err error
		mercariPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatalf("FATAL: Không thể tạo khóa ECDSA cho Mercari: %v", err)
		}

		pubKey := mercariPrivateKey.PublicKey
		jwkMap := map[string]string{
			"crv": "P-256",
			"kty": "EC",
			"x":   base64.RawURLEncoding.EncodeToString(pubKey.X.Bytes()),
			"y":   base64.RawURLEncoding.EncodeToString(pubKey.Y.Bytes()),
		}
		
		mercariJwk, _ = json.Marshal(jwkMap)
		log.Println("Đã tạo cặp khóa DPoP cho Mercari.")
	})
}

func generateDPoP(method, url string) (string, error) {
	initKeys() 

	claims := jwt.MapClaims{
		"iat":  time.Now().Unix(),
		"jti":  uuid.New().String(),
		"htm":  method, 
		"htu":  url,    
		"uuid": mercariUUID,
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("ES256"), claims)

	token.Header["typ"] = "dpop+jwt"
	token.Header["jwk"] = mercariJwk

	return token.SignedString(mercariPrivateKey)
}

func SearchMercariData(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	payload := MercariSearchPayload{
		PageSize:        120,
		PageToken:       "",
		SearchSessionId: uuid.New().String(),
		Source:          "BaseSerp",
		WithItemBrand:   true,
		SearchCondition: MercariSearchCondition{
			Keyword:        req.Keyword,
			ExcludeKeyword: "",
			Sort:           "SORT_SCORE",
			Order:          "ORDER_DESC",
			Status:         []string{},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build search query."})
		return
	}

	searchURL := "https://api.mercari.jp/v2/entities:search"

	httpReq, err := http.NewRequest("POST", searchURL, bytes.NewReader(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API request."})
		return
	}

	dpopToken, err := generateDPoP("POST", searchURL)
	if err != nil {
		log.Printf("Error generating DPoP: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate DPoP token."})
		return
	}

	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:144.0) Gecko/20100101 Firefox/144.0")
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Accept-Language", "ja")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Referer", "https://jp.mercari.com/")
	httpReq.Header.Set("X-Platform", "web")
	httpReq.Header.Set("X-Country-Code", "US")
	httpReq.Header.Set("DPoP", dpopToken)
	httpReq.Header.Set("Origin", "https://jp.mercari.com")
	httpReq.Header.Set("DNT", "1")
	httpReq.Header.Set("Sec-GPC", "1")
	httpReq.Header.Set("Sec-Fetch-Dest", "empty")
	httpReq.Header.Set("Sec-Fetch-Mode", "cors")
	httpReq.Header.Set("Sec-Fetch-Site", "cross-site")
	httpReq.Header.Set("Connection", "keep-alive")
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending request to Mercari: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to contact Mercari API."})
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Mercari response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API response."})
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Mercari API returned non-200 status: %d. Body: %s", resp.StatusCode, string(body))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mercari API returned an error.", "details": string(body)})
		return
	}
	var apiResponse MercariAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Printf("CRITICAL: Failed to unmarshal Mercari JSON. Error: %v", err)
		log.Printf("CRITICAL: Received body from Mercari: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Mercari data."})
		return
	}
	var searchItems []SearchItem
	for _, item := range apiResponse.Items {
		price, _ := strconv.ParseFloat(item.Price, 64)
		var imgURL string
		if len(item.Thumbnails) > 0 {
			imgURL = item.Thumbnails[0]
		}
		searchItems = append(searchItems, SearchItem{
			Name:      item.Name,
			PriceJPY:  price,
			ImageURL:  imgURL,
			ItemURL:   fmt.Sprintf("https://jp.mercari.com/item/%s", item.ID),
			IsSoldOut: (item.Status != "ITEM_STATUS_ON_SALE"),
		})
	}
	c.JSON(http.StatusOK, searchItems)
}

func FetchMercariData(c *gin.Context) {
	var req FetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if !strings.Contains(req.URL, "jp.mercari.com/item/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Mercari URL. Must be a product page (jp.mercari.com/item/)"})
		return
	}

	itemIDMatch := itemIDRegex.FindStringSubmatch(req.URL)
	var itemID string
	if len(itemIDMatch) > 1 {
		itemID = itemIDMatch[1]
	}
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse item ID from URL."})
		return
	}

	apiURL := fmt.Sprintf(
		"https://api.mercari.jp/items/get?id=%s&include_item_attributes=true&include_product_page_component=true&include_non_ui_item_attributes=true&include_donation=true&include_item_attributes_sections=true&include_auction=true&country_code=VN",
		itemID,
	)

	httpReq, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API request."})
		return
	}

	dpopToken, err := generateDPoP("GET", apiURL)
	if err != nil {
		log.Printf("Error generating DPoP: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate DPoP token."})
		return
	}

	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:144.0) Gecko/20100101 Firefox/144.0")
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Accept-Language", "ja")
	httpReq.Header.Set("Referer", "https://jp.mercari.com/")
	httpReq.Header.Set("X-Platform", "web")
	// --- SỬ DỤNG DPoP ĐỘNG ---
	httpReq.Header.Set("DPoP", dpopToken)
	httpReq.Header.Set("Origin", "https://jp.mercari.com")
	httpReq.Header.Set("DNT", "1")
	httpReq.Header.Set("Sec-GPC", "1")
	httpReq.Header.Set("Sec-Fetch-Dest", "empty")
	httpReq.Header.Set("Sec-Fetch-Mode", "cors")
	httpReq.Header.Set("Sec-Fetch-Site", "cross-site")
	httpReq.Header.Set("Connection", "keep-alive")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending request to Mercari: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to contact Mercari API."})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Mercari response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API response."})
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Mercari API (GetItem) returned non-200 status: %d. Body: %s", resp.StatusCode, string(body))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mercari API returned an error.", "details": string(body)})
		return
	}

	var apiResponse MercariItemResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Printf("CRITICAL: Failed to unmarshal Mercari Item JSON. Error: %v", err)
		log.Printf("CRITICAL: Received body from Mercari: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Mercari item data."})
		return
	}

	itemData := apiResponse.Data
	price := itemData.Price

	var imageURLs []string
	for _, photoURL := range itemData.Photos {
		imageURLs = append(imageURLs, photoURL)
	}
	if len(imageURLs) == 0 {
		imageURLs = append(imageURLs, "https://placehold.co/600x600?text=No+Image")
	}

	var condition string = "Không rõ"
	for _, attr := range itemData.ItemAttributes {
		if attr.Name == "商品の状態" {
			condition = attr.Value
			break
		}
	}

	response := FetchResponse{
		Name:          itemData.Name,
		PriceJPY:      price,
		Description:   itemData.Description,
		Condition:     condition,
		ImageURLs:     imageURLs,
		MercariItemID: itemData.ID,
	}

	c.JSON(http.StatusOK, response)
}