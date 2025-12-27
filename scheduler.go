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
func StartReminderScheduler() {
	log.Println("Starting reminder scheduler...")

	// Load configuration from environment variables
	reminderConfig := getReminderConfig()
	log.Printf("Reminder schedule: Daily reminder at %s, Follow-up at %s",
		reminderConfig.ReminderTime, reminderConfig.FollowUpTime)

	// Run scheduler every minute to check if it's time to send reminders
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			currentTime := now.Format("15:04")

			// Check if it's time to send daily reminders
			if currentTime == reminderConfig.ReminderTime {
				log.Println("Time to send daily reminders")
				sendDailyReminders()
			}

			// Check if it's time to send follow-ups
			if currentTime == reminderConfig.FollowUpTime {
				log.Println("Time to send follow-up reminders")
				sendFollowUpReminders()
			}

			// Check for ignored reminders (no response after follow-up time + 2 hours)
			if now.Hour() == 23 && now.Minute() == 0 {
				log.Println("Checking for ignored reminders")
				markIgnoredReminders()
			}
		}
	}
}

// sendDailyReminders sends reminders to all active users
func sendDailyReminders() {
	users, err := GetAllActiveUsers()
	if err != nil {
		log.Printf("Error getting active users: %v", err)
		return
	}

	today := time.Now().Format("2006-01-02")
	successCount := 0
	failCount := 0

	for _, user := range users {
		// Check if reminder already sent today
		existingLog, err := GetDailyLogByUserAndDate(user.ID, today)
		if err == nil && existingLog != nil {
			log.Printf("Reminder already sent to user %s today", user.Name)
			continue
		}

		// Create daily log entry
		_, err = CreateDailyLog(user.ID, today)
		if err != nil {
			log.Printf("Error creating daily log for user %s: %v", user.Name, err)
			failCount++
			continue
		}

		// Send reminder message
		message := fmt.Sprintf("Good morning, %s! 🌅\n\nDaily reminder: %s\n\nReply 'Yes' or 'Done' when you complete it!",
			user.Name, user.Task)

		if err := sendTwilioWhatsApp(user.PhoneNumber, message); err != nil {
			log.Printf("Failed to send reminder to %s: %v", user.Name, err)
			failCount++
		} else {
			log.Printf("Reminder sent to %s (%s)", user.Name, user.PhoneNumber)
			successCount++
		}

		// Small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Daily reminders sent: %d successful, %d failed", successCount, failCount)
}

// sendFollowUpReminders sends follow-up reminders to users who haven't responded
func sendFollowUpReminders() {
	today := time.Now().Format("2006-01-02")

	// Get all pending logs for today
	pendingLogs, err := GetPendingLogsForDate(today)
	if err != nil {
		log.Printf("Error getting pending logs: %v", err)
		return
	}

	successCount := 0
	failCount := 0

	for _, logEntry := range pendingLogs {
		// Check if follow-up already sent
		if logEntry.FollowUpSentAt != nil {
			continue
		}

		// Get user info by ID
		user, err := GetUserByID(logEntry.UserID)
		if err != nil {
			log.Printf("User not found for log ID %d: %v", logEntry.ID, err)
			failCount++
			continue
		}

		// Send follow-up message
		message := fmt.Sprintf("Hi %s, just checking in! 👋\n\nHave you completed your task: %s?\n\nReply 'Yes' or 'No'",
			user.Name, user.Task)

		if err := sendTwilioWhatsApp(user.PhoneNumber, message); err != nil {
			log.Printf("Failed to send follow-up to %s: %v", user.Name, err)
			failCount++
		} else {
			// Mark follow-up as sent
			if err := MarkFollowUpSent(user.ID, today); err != nil {
				log.Printf("Error marking follow-up sent: %v", err)
			}
			log.Printf("Follow-up sent to %s (%s)", user.Name, user.PhoneNumber)
			successCount++
		}

		// Small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Follow-up reminders sent: %d successful, %d failed", successCount, failCount)
}

// markIgnoredReminders marks reminders as ignored if no response by end of day
// and notifies guilt buddies
func markIgnoredReminders() {
	today := time.Now().Format("2006-01-02")

	// Get all pending logs for today
	pendingLogs, err := GetPendingLogsForDate(today)
	if err != nil {
		log.Printf("Error getting pending logs: %v", err)
		return
	}

	guiltBuddyNotificationsSent := 0
	guiltBuddyNotificationsFailed := 0

	for _, logEntry := range pendingLogs {
		// Get user info
		user, err := GetUserByID(logEntry.UserID)
		if err != nil {
			log.Printf("Error getting user for log ID %d: %v", logEntry.ID, err)
			continue
		}

		// Mark as ignored
		if err := MarkLogAsIgnored(logEntry.UserID, today); err != nil {
			log.Printf("Error marking log as ignored: %v", err)
			continue
		}
		log.Printf("Marked log as ignored for user %s (ID: %d)", user.Name, logEntry.UserID)

		// Send notification to guilt buddy if configured
		if user.GuiltBuddyPhone != "" && user.GuiltBuddyName != "" {
			guiltBuddyMsg := fmt.Sprintf("Hey %s! 😔\n\n%s didn't complete their task today: %s\n\nMaybe check in with them?",
				user.GuiltBuddyName, user.Name, user.Task)

			if err := sendTwilioWhatsApp(user.GuiltBuddyPhone, guiltBuddyMsg); err != nil {
				log.Printf("Failed to send guilt buddy notification to %s: %v", user.GuiltBuddyName, err)
				guiltBuddyNotificationsFailed++
			} else {
				log.Printf("Guilt buddy notification sent to %s for user %s", user.GuiltBuddyName, user.Name)
				guiltBuddyNotificationsSent++
			}

			// Small delay to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Printf("Marked %d reminders as ignored", len(pendingLogs))
	if guiltBuddyNotificationsSent > 0 || guiltBuddyNotificationsFailed > 0 {
		log.Printf("Guilt buddy notifications: %d sent, %d failed", guiltBuddyNotificationsSent, guiltBuddyNotificationsFailed)
	}
}
