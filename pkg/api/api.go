package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"product_management/pkg/cache"
	"product_management/pkg/models"
	"product_management/pkg/rabbitmq"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProductById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
		return
	}

	cacheKey := fmt.Sprintf("product_%d", id)
	cached, err := cache.Get(cacheKey)
	if err == nil {
		log.Printf("Cache hit for product ID %d: %s", id, cached)
	} else {
		log.Printf("Cache miss for product ID %d", id)
	}

	product, err := models.GetProductByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	jsonData, _ := json.Marshal(product)
	log.Printf("Setting cache for product ID %d: %s", id, jsonData)
	cache.Set(cacheKey, string(jsonData))                               // Adjust to match the Set method signature.
	log.Printf("Database response for product ID %d: %+v", id, product) // Log database response
	c.JSON(http.StatusOK, product)
}

func GetProducts(c *gin.Context) {
	products, err := models.GetAllProducts(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func AddProduct(c *gin.Context) {
	var newProduct models.Product

	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := models.SaveProduct(context.Background(), &newProduct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save product"})
		return
	}

	// Publish product images to RabbitMQ
	imageData, _ := json.Marshal(newProduct.ProductImages)
	rabbitmq.PublishMessage(string(imageData))

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully", "product_id": newProduct.ID})
}

func UpdateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := models.UpdateProduct(context.Background(), &product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update product"})
		return
	}

	cacheKey := fmt.Sprintf("product_%d", product.ID)
	cache.Delete(cacheKey) // Invalidate cache
	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}
