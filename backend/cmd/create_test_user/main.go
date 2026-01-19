package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type TestUser struct {
	Email    string
	Username string
}

func main() {
	ctx := context.Background()

	// Test users to create
	users := []TestUser{
		{Email: "test@mail.com", Username: "testuser"},
		{Email: "buyer@mail.com", Username: "buyer"},
	}

	// Hash the password
	password := "Test123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatal("Error hashing password:", err)
	}
	fmt.Println("Generated hash:", string(hash))

	// Connect to database
	connStr := "postgres://auction:auction123@localhost:5432/auction_db?sslmode=disable"
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer conn.Close(ctx)

	// Create each test user
	query := `
		INSERT INTO users (email, username, password_hash, email_verified, role)
		VALUES ($1, $2, $3, true, 'user')
		ON CONFLICT (email) DO UPDATE SET password_hash = $3, email_verified = true
		RETURNING id, email, username
	`

	for _, user := range users {
		var id, email, username string
		err = conn.QueryRow(ctx, query, user.Email, user.Username, string(hash)).Scan(&id, &email, &username)
		if err != nil {
			log.Printf("Error inserting user %s: %v", user.Email, err)
			continue
		}
		fmt.Printf("User created/updated: ID=%s, Email=%s, Username=%s\n", id, email, username)
	}

	// Verify the password works for first user
	var storedHash string
	err = conn.QueryRow(ctx, "SELECT password_hash FROM users WHERE email = $1", "test@mail.com").Scan(&storedHash)
	if err != nil {
		log.Fatal("Error fetching user:", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		fmt.Println("Password verification FAILED:", err)
		os.Exit(1)
	}
	fmt.Println("Password verification PASSED!")
}
