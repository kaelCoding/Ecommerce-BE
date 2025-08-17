package router

import (
	"html/template"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/handlers"
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

func SetupRouter() *gin.Engine {
	r := gin.Default()
	db := database.DB

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:5173",
		"https://kaelcoding.github.io/Ecommerce-FE",
		"https://ecommerce-fe-nu.vercel.app",
		"https://ecommerce-fe-git-main-kaels-projects-7e91b725.vercel.app",
		"https://ecommerce-au20xy41z-kaels-projects-7e91b725.vercel.app",
		"https://tunitoku.netlify.app",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	r.GET("/", handler)

	api := r.Group("/api/v1")
	{
		// --- Public Routes (No authentication required) ---
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterUser(db))
			auth.POST("/login", handlers.LoginUser(db))
		}

		// Public product routes
		api.GET("/products", handlers.GetProducts)
		api.GET("/products/:id", handlers.GetProductByID)
		api.GET("/products/search", handlers.SearchProducts)

		// Public category routes
		api.GET("/categories", handlers.GetCategory) 
		api.GET("/categories/:id", handlers.GetCategoryByID)
		api.GET("/categories/:id/products", handlers.GetProductsByCategory)

		// Public feedback route
		api.POST("/feedback", handlers.SendFeedbackHandler)

		// --- Protected Routes (Requires user to be logged in) ---
		protected := api.Group("/")
		protected.Use(handlers.AuthMiddleware())
		{
			protected.GET("/profile", handlers.GetUser(db))
			protected.POST("/orders", handlers.CreateOrderHandler)
		}

		admin := api.Group("/admin")
		admin.Use(handlers.AuthMiddleware())
		admin.Use(handlers.AdminOnlyMiddleware())
		{
			// User management
			admin.GET("/users", handlers.GetAllUsers(db))

			admin.POST("/products", handlers.AddProduct)
			admin.PUT("/products/:id", handlers.UpdateProduct)
			admin.DELETE("/products/:id", handlers.DeleteProduct)

			// Category management
			admin.POST("/categories", handlers.AddCategory)
			admin.PUT("/categories/:id", handlers.UpdateCategory)
			admin.DELETE("/categories/:id", handlers.DeleteCategory)
		}
	}

	return r
}
