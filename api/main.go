package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get database path from environment variable or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./reminder_system.db"
	}

	// Initialize SQLite database
	if err := InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// Start the reminder scheduler in a goroutine
	go StartReminderScheduler()

	// Register HTTP handlers
	// New endpoints for user management and reminders
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/webhook/twilio", twilioWebhookHandler)
	http.HandleFunc("/user", getUserHandler)
	http.HandleFunc("/user/logs", getUserLogsHandler)
	http.HandleFunc("/users", getAllUsersHandler)

	// Serve static files (frontend)
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Get server port from environment variable or use default
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	fmt.Printf("starting server at port %s\n", serverPort)
	fmt.Println("Endpoints available:")
	fmt.Println("  GET  / - Dashboard (Frontend)")
	fmt.Println("  POST /register - Register a new user")
	fmt.Println("  POST /webhook/twilio - Twilio webhook for incoming messages")
	fmt.Println("  GET  /user?phone=<phone> - Get user info")
	fmt.Println("  GET  /user/logs?phone=<phone> - Get user daily logs")
	fmt.Println("  GET  /users - Get all active users")

	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}
