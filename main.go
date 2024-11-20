package main

import (
	"log"
	"product_management/pkg/api"
	"product_management/pkg/cache"
	"product_management/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	cache.InitCache()
	router := gin.Default()

	router.GET("/product/:id", api.GetProductById)
	router.GET("/product", api.GetProducts)
	router.POST("/product", api.AddProduct)
	router.PUT("/product/:id", api.UpdateProduct)

	// Start the server
	log.Println("Server starting on port 3000...")
	if err := router.Run(":3000"); err != nil {
		log.Fatalf("Eror starting the server: %v", err)
	}
}
