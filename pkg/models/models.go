package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"product_management/pkg/database"
	"strings"
)

type Product struct {
	ID                 int      `json:"id"`
	UserID             int      `json:"user_id"`
	ProductName        string   `json:"product_name"`
	ProductDescription string   `json:"product_description"`
	ProductImages      []string `json:"product_images"`
	ProductPrice       float64  `json:"product_price"`
	CompressedImages   []string `json:"compressed_images"`
}

func GetProductByID(ctx context.Context, id int) (*Product, error) {
	query := `
		SELECT id, user_id, product_name, product_description, product_images, product_price, compressed_product_images
		FROM products WHERE id = $1
	`
	row := database.DB.QueryRow(ctx, query, id)

	var product Product
	var productImages []string
	var compressedImages []byte // Assuming this is not an array of text

	err := row.Scan(
		&product.ID,
		&product.UserID,
		&product.ProductName,
		&product.ProductDescription,
		&productImages, // Scan into []string
		&product.ProductPrice,
		&compressedImages,
	)
	if err != nil {
		log.Printf("Error scanning row: %v", err) // Log any scanning errors
		return nil, err
	}

	// Assign the scanned images to the product struct
	product.ProductImages = productImages

	// If compressed images are stored as text, handle them similarly as needed
	if err := json.Unmarshal(compressedImages, &product.CompressedImages); err != nil {
		log.Printf("Error unmarshaling compressed images: %v", err)
	}

	log.Printf("Fetched product: %+v", product) // Log the fetched product

	return &product, nil
}

func GetAllProducts(ctx context.Context) ([]Product, error) {
	query := `
		SELECT id, user_id, product_name, product_description, product_images, product_price, compressed_product_images
		FROM products
	`
	rows, err := database.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var productImages, compressedImages []byte
		var productImagesNull, compressedImagesNull sql.NullString // For handling NULLs in the database

		// Scan the row into the product struct
		err := rows.Scan(
			&product.ID,
			&product.UserID,
			&product.ProductName,
			&product.ProductDescription,
			&productImagesNull,
			&product.ProductPrice,
			&compressedImagesNull,
		)
		if err != nil {
			return nil, err
		}

		// Handle NULL values by checking the sql.NullString
		if productImagesNull.Valid {
			productImages = []byte(productImagesNull.String)
		} else {
			productImages = nil
		}

		if compressedImagesNull.Valid {
			compressedImages = []byte(compressedImagesNull.String)
		} else {
			compressedImages = nil
		}

		// Log the raw values for debugging
		log.Printf("Raw productImages: %s", string(productImages))
		log.Printf("Raw compressedImages: %s", string(compressedImages))

		// Handle product images (delimited string)
		if productImages != nil && len(productImages) > 0 {
			// Split by the delimiter (e.g., §)
			product.ProductImages = strings.Split(string(productImages), "§")
		}

		// Handle compressed images (if needed)
		if compressedImages != nil && len(compressedImages) > 0 {
			// If you have a specific logic for compressed images, implement it here
			product.CompressedImages = strings.Split(string(compressedImages), "§")
		}

		// Append the product to the list
		products = append(products, product)
	}

	// Check if any error occurred during the iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func SaveProduct(ctx context.Context, product *Product) error {
	// Convert product images slice to PostgreSQL array format
	productImages := fmt.Sprintf("{%s}", strings.Join(product.ProductImages, ","))

	// Log the product data before saving
	log.Printf("Inserting product: %+v", product)

	// Prepare the insert query
	query := `
		INSERT INTO products (user_id, product_name, product_description, product_images, product_price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	// Execute the query
	err := database.DB.QueryRow(ctx, query, product.UserID, product.ProductName, product.ProductDescription, productImages, product.ProductPrice).Scan(&product.ID)
	if err != nil {
		log.Printf("Failed to execute query: %v", err) // Log the error
		return fmt.Errorf("failed to save product: %v", err)
	}

	return nil
}

func UpdateProduct(ctx context.Context, product *Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, product_images = $4, compressed_product_images = $5
		WHERE id = $6
	`

	_, err := database.DB.Exec(ctx, query,
		product.ProductName,
		product.ProductDescription,
		product.ProductPrice,
		product.ProductImages,
		product.CompressedImages,
		product.ID,
	)
	if err != nil {
		log.Printf("Failed to update product with ID %d: %v", product.ID, err)
		return err
	}
	return nil
}
