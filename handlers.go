package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	PhoneNumber     string `json:"phone_number"`
	Name            string `json:"name"`
	Task            string `json:"task"`
	GuiltBuddyPhone string `json:"guilt_buddy_phone"` // Optional: Phone number to notify when task is not completed
	GuiltBuddyName  string `json:"guilt_buddy_name"`  // Optional: Name of the guilt buddy
}

// RegisterResponse represents the user registration response
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}

// registerHandler handles user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate input
	if req.PhoneNumber == "" || req.Name == "" || req.Task == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Phone number, name, and task are required",
		})
		return
	}

	// Normalize phone number (ensure it starts with +)
	if !strings.HasPrefix(req.PhoneNumber, "+") {
		req.PhoneNumber = "+" + req.PhoneNumber
	}

	// Normalize guilt buddy phone number if provided
	if req.GuiltBuddyPhone != "" && !strings.HasPrefix(req.GuiltBuddyPhone, "+") {
		req.GuiltBuddyPhone = "+" + req.GuiltBuddyPhone
	}

	// Check if user already exists
	existingUser, err := GetUserByPhone(req.PhoneNumber)
	if err == nil && existingUser != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "User with this phone number already exists",
			User:    existingUser,
		})
		return
	}

	// Create new user
	user, err := CreateUser(req.PhoneNumber, req.Name, req.Task, req.GuiltBuddyPhone, req.GuiltBuddyName)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Failed to create user",
		})
		return
	}

	// Send welcome message
	welcomeMsg := fmt.Sprintf("Welcome %s! You've been registered for daily reminders about: %s\n\nYou'll receive a daily reminder. Reply with 'Yes' or 'Done' to confirm completion.",
		user.Name, user.Task)
	if user.GuiltBuddyPhone != "" && user.GuiltBuddyName != "" {
		welcomeMsg += fmt.Sprintf("\n\nIf you don't complete your task, %s will be notified.", user.GuiltBuddyName)
	}
	if err := sendTwilioWhatsApp(user.PhoneNumber, welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
	}

	// Send notification to guilt buddy if provided
	if user.GuiltBuddyPhone != "" && user.GuiltBuddyName != "" {
		guiltBuddyMsg := fmt.Sprintf("Hi %s! 👋\n\n%s has added you as their accountability buddy for: %s\n\nYou'll receive a notification if they don't complete their daily task.",
			user.GuiltBuddyName, user.Name, user.Task)
		if err := sendTwilioWhatsApp(user.GuiltBuddyPhone, guiltBuddyMsg); err != nil {
			log.Printf("Failed to send guilt buddy notification: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RegisterResponse{
		Success: true,
		Message: "User registered successfully",
		User:    user,
	})
}

// twilioWebhookHandler handles incoming WhatsApp messages from Twilio
func twilioWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data from Twilio
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Extract relevant fields
	from := r.FormValue("From") // e.g., "whatsapp:+919999999999"
	body := r.FormValue("Body") // The message text

	// Remove "whatsapp:" prefix
	phoneNumber := strings.TrimPrefix(from, "whatsapp:")

	log.Printf("Received message from %s: %s", phoneNumber, body)

	// Get user by phone number
	user, err := GetUserByPhone(phoneNumber)
	if err != nil {
		log.Printf("User not found for phone %s: %v", phoneNumber, err)
		// Send a message to register first
		sendTwilioWhatsApp(phoneNumber, "You're not registered yet. Please register first to use this service.")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get today's date
	today := time.Now().Format("2006-01-02")

	// Check if there's a pending log for today
	_, err = GetDailyLogByUserAndDate(user.ID, today)
	if err != nil {
		if err == sql.ErrNoRows {
			// No log for today, inform user
			sendTwilioWhatsApp(phoneNumber, "No reminder was sent today. You'll receive your next reminder as scheduled.")
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Printf("Error getting daily log: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Process the response
	bodyLower := strings.ToLower(strings.TrimSpace(body))
	var responseMsg string

	if bodyLower == "yes" || bodyLower == "done" || bodyLower == "completed" || bodyLower == "y" {
		// User confirmed completion
		if err := UpdateDailyLogResponse(user.ID, today, body); err != nil {
			log.Printf("Error updating daily log: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseMsg = fmt.Sprintf("Great job, %s! ✅ Your task '%s' has been marked as completed for today.", user.Name, user.Task)
	} else if bodyLower == "no" || bodyLower == "not yet" || bodyLower == "n" {
		// User said no
		if err := UpdateDailyLogResponse(user.ID, today, body); err != nil {
			log.Printf("Error updating daily log: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseMsg = fmt.Sprintf("No worries, %s. Remember to complete your task: %s", user.Name, user.Task)
	} else {
		// Unknown response
		responseMsg = "Please reply with 'Yes' or 'No' to confirm your task completion."
	}

	// Send response back to user
	if err := sendTwilioWhatsApp(phoneNumber, responseMsg); err != nil {
		log.Printf("Failed to send response message: %v", err)
	}

	// Respond to Twilio with 200 OK
	w.WriteHeader(http.StatusOK)
}

// getUserHandler retrieves user information by phone number
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	phoneNumber := r.URL.Query().Get("phone")
	if phoneNumber == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	// Normalize phone number
	if !strings.HasPrefix(phoneNumber, "+") {
		phoneNumber = "+" + phoneNumber
	}

	user, err := GetUserByPhone(phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// getUserLogsHandler retrieves daily logs for a user
func getUserLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	phoneNumber := r.URL.Query().Get("phone")
	if phoneNumber == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	// Normalize phone number
	if !strings.HasPrefix(phoneNumber, "+") {
		phoneNumber = "+" + phoneNumber
	}

	user, err := GetUserByPhone(phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get logs (limit to last 30 days)
	logs, err := GetUserDailyLogs(user.ID, 30)
	if err != nil {
		log.Printf("Error getting user logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
		"logs": logs,
	})
}

// getAllUsersHandler retrieves all active users
func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := GetAllActiveUsers()
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(users),
		"users": users,
	})
}
