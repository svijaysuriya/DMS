package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// ReminderConfig holds configuration for reminder scheduling
type ReminderConfig struct {
	ReminderTime  string // Time to send daily reminder (e.g., "09:00")
	FollowUpTime  string // Time to send follow-up (e.g., "18:00")
	FollowUpDelay int    // Hours after reminder to send follow-up
}

// getReminderConfig loads reminder configuration from environment variables
func getReminderConfig() ReminderConfig {
	reminderTime := os.Getenv("REMINDER_TIME")
	if reminderTime == "" {
		reminderTime = "09:00" // Default to 9 AM
	}

	followUpTime := os.Getenv("FOLLOWUP_TIME")
	if followUpTime == "" {
		followUpTime = "18:00" // Default to 6 PM
	}

	return ReminderConfig{
		ReminderTime:  reminderTime,
		FollowUpTime:  followUpTime,
		FollowUpDelay: 6, // 6 hours after reminder
	}
}

// StartReminderScheduler starts the background scheduler for sending reminders
// It checks every minute and processes each user based on their local timezone
func StartReminderScheduler() {
	log.Println("Starting reminder scheduler with per-user timezone support...")

	// Load configuration from environment variables
	reminderConfig := getReminderConfig()
	log.Printf("Reminder schedule: Daily reminder at %s (user's local time), Follow-up at %s (user's local time)",
		reminderConfig.ReminderTime, reminderConfig.FollowUpTime)

	// Run scheduler every minute to check each user's local time
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Process all users based on their individual timezones
			processUsersForReminders(reminderConfig)
		}
	}
}

// processUsersForReminders checks each user's local time and sends appropriate reminders
func processUsersForReminders(config ReminderConfig) {
	users, err := GetAllActiveUsers()
	if err != nil {
		log.Printf("Error getting active users: %v", err)
		return
	}

	for _, user := range users {
		// Get user's local time
		userLocalTime := GetUserLocalTime(user.Timezone)
		userCurrentTime := userLocalTime.Format("15:04")
		userToday := userLocalTime.Format("2006-01-02")

		// Check if it's reminder time for this user
		if userCurrentTime == config.ReminderTime {
			sendReminderToUser(user, userToday)
		}

		// Check if it's follow-up time for this user
		if userCurrentTime == config.FollowUpTime {
			sendFollowUpToUser(user, userToday)
		}

		// Check if it's end of day (23:00) for this user - mark ignored
		if userLocalTime.Hour() == 23 && userLocalTime.Minute() == 0 {
			markUserReminderAsIgnored(user, userToday)
		}
	}
}

// sendReminderToUser sends a daily reminder to a specific user
func sendReminderToUser(user User, userToday string) {
	// Check if reminder already sent today (in user's timezone)
	existingLog, err := GetDailyLogByUserAndDate(user.ID, userToday)
	if err == nil && existingLog != nil {
		// Reminder already sent today
		return
	}

	// Create daily log entry
	_, err = CreateDailyLog(user.ID, userToday)
	if err != nil {
		log.Printf("Error creating daily log for user %s: %v", user.Name, err)
		return
	}

	// Send reminder message
	message := fmt.Sprintf("Good morning, %s! 🌅\n\nDaily reminder: %s\n\nReply 'Yes' or 'Done' when you complete it!",
		user.Name, user.Task)

	if err := sendTwilioWhatsApp(user.PhoneNumber, message); err != nil {
		log.Printf("Failed to send reminder to %s: %v", user.Name, err)
	} else {
		log.Printf("Reminder sent to %s (%s) [TZ: %s]", user.Name, user.PhoneNumber, user.Timezone)
	}
}

// sendFollowUpToUser sends a follow-up reminder to a specific user if they haven't responded
func sendFollowUpToUser(user User, userToday string) {
	// Check if there's a pending log for today (in user's timezone)
	logEntry, err := GetDailyLogByUserAndDate(user.ID, userToday)
	if err != nil {
		// No log for today, nothing to follow up on
		return
	}

	// Check if follow-up already sent or already responded
	if logEntry.FollowUpSentAt != nil || logEntry.Status != "pending" {
		return
	}

	// Send follow-up message
	message := fmt.Sprintf("Hi %s, just checking in! 👋\n\nHave you completed your task: %s?\n\nReply 'Yes' or 'No'",
		user.Name, user.Task)

	if err := sendTwilioWhatsApp(user.PhoneNumber, message); err != nil {
		log.Printf("Failed to send follow-up to %s: %v", user.Name, err)
	} else {
		// Mark follow-up as sent
		if err := MarkFollowUpSent(user.ID, userToday); err != nil {
			log.Printf("Error marking follow-up sent: %v", err)
		}
		log.Printf("Follow-up sent to %s (%s) [TZ: %s]", user.Name, user.PhoneNumber, user.Timezone)
	}
}

// markUserReminderAsIgnored marks a user's reminder as ignored if no response by end of day
// and notifies guilt buddy
func markUserReminderAsIgnored(user User, userToday string) {
	// Check if there's a pending log for today (in user's timezone)
	logEntry, err := GetDailyLogByUserAndDate(user.ID, userToday)
	if err != nil {
		// No log for today
		return
	}

	// Only mark as ignored if still pending
	if logEntry.Status != "pending" {
		return
	}

	// Mark as ignored
	if err := MarkLogAsIgnored(user.ID, userToday); err != nil {
		log.Printf("Error marking log as ignored for %s: %v", user.Name, err)
		return
	}
	log.Printf("Marked log as ignored for user %s [TZ: %s]", user.Name, user.Timezone)

	// Send notification to guilt buddy if configured
	if user.GuiltBuddyPhone != "" && user.GuiltBuddyName != "" {
		guiltBuddyMsg := fmt.Sprintf("Hey %s! 😔\n\n%s didn't complete their task today: %s\n\nMaybe check in with them?",
			user.GuiltBuddyName, user.Name, user.Task)

		if err := sendTwilioWhatsApp(user.GuiltBuddyPhone, guiltBuddyMsg); err != nil {
			log.Printf("Failed to send guilt buddy notification to %s: %v", user.GuiltBuddyName, err)
		} else {
			log.Printf("Guilt buddy notification sent to %s for user %s", user.GuiltBuddyName, user.Name)
		}
	}
}
