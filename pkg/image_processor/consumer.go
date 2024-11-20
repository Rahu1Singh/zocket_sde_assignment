package image_processor

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

var db *pgxpool.Pool

func ConnectDB() {
	var err error
	dsn := "postgres://api_user:secure_password@localhost:5432/product_management"
	db, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	fmt.Println("Connected to database")
}

func compressImage(url string) ([]byte, error) {
	// Download the image
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Compress the image
	compressed := imaging.Resize(img, 800, 0, imaging.Lanczos)
	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, compressed, imaging.JPEG)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

func updateCompressedImages(productID int, compressedURLs []string) {
	query := `
		UPDATE products
		SET compressed_product_images = $1
		WHERE id = $2
	`
	_, err := db.Exec(context.Background(), query, compressedURLs, productID)
	if err != nil {
		log.Printf("Failed to update product with ID %d: %v", productID, err)
	}
}

func ConsumeMessages() {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer channel.Close()

	msgs, err := channel.Consume(
		"image_processing", // Queue
		"",                 // Consumer
		true,               // Auto-Ack
		false,              // Exclusive
		false,              // No-local
		false,              // No-wait
		nil,                // Args
	)
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	// Process messages
	for msg := range msgs {
		imageURLs := strings.Split(string(msg.Body), ",")
		var compressedURLs []string

		for _, url := range imageURLs {
			compressed, err := compressImage(url)
			if err != nil {
				log.Printf("Failed to process image %s: %v", url, err)
				continue
			}

			// Simulate saving compressed image to S3 (just saving locally for now)
			filename := fmt.Sprintf("compressed_%s.jpg", strings.Split(url, "/")[len(url)-1])
			os.WriteFile(filename, compressed, 0644)
			compressedURLs = append(compressedURLs, filename)
		}

		// Update database with compressed image URLs (replace `productID` with actual logic)
		updateCompressedImages(1, compressedURLs)
	}
}
