// env
package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads the .env file if it exists
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// LoadFacebookConfig loads facebook business api configs
func LoadFacebookConfig() {
	LoadEnv()
	token := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("FACEBOOK_ACCESS_TOKEN is not set in the environment")
	}
}
