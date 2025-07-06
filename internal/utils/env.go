// env
package utils

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv loads the .env file if it exists
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}
