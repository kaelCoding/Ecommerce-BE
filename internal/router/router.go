package router

import (
    "net/http"
    "html/template"

    "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/handlers"
	"github.com/kaelCoding/toyBE/internal/database"
)

type Data struct {
    Name string
}

func handler(c *gin.Context) {
    data := Data{Name: "Thế giới"}
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
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{
        "http://localhost:5173", 
        "https://lbltxmh1-5173.asse.devtunnels.ms",
        "https://ecommerce-fe-4ezt.onrender.com",
        "https://kaelcoding.github.io/Ecommerce-FE"
    } 
    config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
    config.AllowCredentials = true 
    r.Use(cors.New(config))

    r.Static("/uploads", "./uploads") // Serve static files

    r.GET("/", handler)
	// User routes
	r.POST("/register", handlers.RegisterUser(database.DB))
	r.POST("/login", handlers.LoginUser(database.DB))

    r.GET("/auth/info", handlers.GetUser(database.DB))
    r.GET("/users", handlers.GetAllUsers(database.DB))

	// Product routes
	r.POST("/products", handlers.AddProduct) 
	r.GET("/products", handlers.GetProducts)
	r.GET("/products/:id", handlers.GetProductByID)
	r.PUT("/products/:id", handlers.UpdateProduct) 
	r.DELETE("/products/:id", handlers.DeleteProduct)

	// Category routes
	r.POST("/category", handlers.AddCategory)
    r.GET("/category", handlers.GetCategory)
	r.GET("/category/:id", handlers.GetCategoryByID)
	r.PUT("/category/:id", handlers.UpdateCategory)
	r.DELETE("/category/:id", handlers.DeleteCategory)

    r.POST("/orders", handlers.CreateOrderHandler)

    return r
}