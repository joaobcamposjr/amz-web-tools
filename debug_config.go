package main

import (
	"log"
	"os"

	"amz-web-tools/backend/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	log.Printf("🔧 Configuration Debug:")
	log.Printf("  DB Host: %s", cfg.DBHost)
	log.Printf("  DB Port: %s", cfg.DBPort)
	log.Printf("  DB Name: %s", cfg.DBName)
	log.Printf("  Plate API URL: %s", cfg.PlateAPIURL)
	log.Printf("  Plate API Key: %s", cfg.PlateAPIKey)
	log.Printf("  Environment: %s", cfg.Environment)

	// Check if .env file exists
	if _, err := os.Stat(".env"); err == nil {
		log.Printf("✅ .env file found")
	} else {
		log.Printf("❌ .env file not found")
	}

	// Check environment variables
	log.Printf("🌍 Environment Variables:")
	log.Printf("  PLATE_API_URL: %s", os.Getenv("PLATE_API_URL"))
	log.Printf("  PLATE_API_KEY: %s", os.Getenv("PLATE_API_KEY"))
}

