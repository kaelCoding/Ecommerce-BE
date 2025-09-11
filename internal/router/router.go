package router

import (
	"html/template"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/handlers"
	"github.com/kaelCoding/toyBE/internal/chat"
)

type Data struct {
	Name string
}

func handler(c *gin.Context) {
	data := Data{Name: "Hello world!"}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func SetupRouter(hub *chat.Hub) *gin.Engine {
	r := gin.Default()
	db := database.DB

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:5173",
		"https://tunitoku.netlify.app",
		"https://tunitoku.store",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	r.GET("/", handler)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterUser(db))
			auth.POST("/login", handlers.LoginUser(db))
		}

		api.GET("/products", handlers.GetProducts)
		api.GET("/products/ids", handlers.GetAllProductIDs)
		api.GET("/products/:id", handlers.GetProductByID)
		api.GET("/products/search", handlers.SearchProducts)

		api.GET("/categories", handlers.GetCategory) 
		api.GET("/categories/:id", handlers.GetCategoryByID)
		api.GET("/categories/:id/products", handlers.GetProductsByCategory)
		api.GET("/categories/:id/products/limit", handlers.GetProductsByCategoryIDWithLimit)

		api.POST("/feedback", handlers.SendFeedbackHandler)
		api.POST("/spin", handlers.SpinByShippingCode(db))
		api.GET("/rewards", handlers.GetRewards(db))

		protected := api.Group("/")
		protected.Use(handlers.AuthMiddleware())
		{
			protected.GET("/profile", handlers.GetUser(db))
			protected.POST("/orders", handlers.CreateOrderHandler)
			protected.GET("/ws", handlers.ChatEndpoint(hub, db))
            protected.GET("/chat/history", handlers.GetChatHistory(db))
			protected.GET("/admin-info", handlers.GetAdminInfo(db))
			protected.POST("/fcm/token", handlers.UpdateFCMToken(db)) 
		}

		admin := api.Group("/admin")
		admin.Use(handlers.AuthMiddleware())
		admin.Use(handlers.AdminOnlyMiddleware())
		{
			admin.GET("/users", handlers.GetAllUsers(db))

			admin.POST("/products", handlers.AddProduct)
			admin.PUT("/products/:id", handlers.UpdateProduct)
			admin.DELETE("/products/:id", handlers.DeleteProduct)

			admin.POST("/categories", handlers.AddCategory)
			admin.PUT("/categories/:id", handlers.UpdateCategory)
			admin.DELETE("/categories/:id", handlers.DeleteCategory)

			admin.GET("/orders", handlers.GetAllOrders(db))
    		admin.PUT("/orders/:id/shipping-code", handlers.UpdateShippingCode(db))
			admin.POST("/rewards", handlers.AddReward(db))
			admin.PUT("/rewards/:id", handlers.UpdateReward(db))
			admin.DELETE("/rewards/:id", handlers.DeleteReward(db))
		}
	}

	return r
}