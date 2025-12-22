package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("../.env")

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secret-jwt-key-change-in-production"
	}

	userID := "test-user-123"
	email := "test@example.com"
	role := 1

	if len(os.Args) > 1 {
		userID = os.Args[1]
	}
	if len(os.Args) > 2 {
		email = os.Args[2]
	}

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== JWT Token Generated ===")
	fmt.Println()
	fmt.Println(tokenString)
	fmt.Println()
	fmt.Println("=== Token Claims ===")
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Role: %d\n", role)
	fmt.Printf("Expires: %s\n", time.Now().Add(24*time.Hour).Format(time.RFC3339))
}
