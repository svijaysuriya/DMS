package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// isColumnExistsError checks if the error is due to column already existing
func isColumnExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate column name")
}

// User represents a registered user
type User struct {
	ID              int64     `json:"id"`
	PhoneNumber     string    `json:"phone_number"`
	Name            string    `json:"name"`
	Task            string    `json:"task"`
	GuiltBuddyPhone string    `json:"guilt_buddy_phone"` // Phone number to notify when task is not completed
	GuiltBuddyName  string    `json:"guilt_buddy_name"`  // Name of the guilt buddy
	CreatedAt       time.Time `json:"created_at"`
	Active          bool      `json:"active"`
}

// DailyLog represents a daily reminder and response log
type DailyLog struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	Date           string     `json:"date"` // YYYY-MM-DD format
	ReminderSentAt time.Time  `json:"reminder_sent_at"`
	Response       string     `json:"response"`          // "yes", "no", "ignored", etc.
	RespondedAt    *time.Time `json:"responded_at"`      // NULL if not responded
	FollowUpSentAt *time.Time `json:"follow_up_sent_at"` // NULL if no follow-up sent
	Status         string     `json:"status"`            // "pending", "completed", "missed", "ignored"
}

// InitDB initializes the SQLite database and creates tables
func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		phone_number TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		task TEXT NOT NULL,
		guilt_buddy_phone TEXT DEFAULT '',
		guilt_buddy_name TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		active BOOLEAN DEFAULT 1
	);`

	if _, err = db.Exec(createUsersTable); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Add guilt buddy columns if they don't exist (for existing databases)
	alterTableQueries := []string{
		`ALTER TABLE users ADD COLUMN guilt_buddy_phone TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN guilt_buddy_name TEXT DEFAULT ''`,
	}

	for _, query := range alterTableQueries {
		_, err = db.Exec(query)
		// Ignore error if column already exists
		if err != nil && !isColumnExistsError(err) {
			log.Printf("Warning: Could not add column: %v", err)
		}
	}

	// Create daily_logs table
	createDailyLogsTable := `
	CREATE TABLE IF NOT EXISTS daily_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		reminder_sent_at DATETIME NOT NULL,
		response TEXT DEFAULT '',
		responded_at DATETIME,
		follow_up_sent_at DATETIME,
		status TEXT DEFAULT 'pending',
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(user_id, date)
	);`

	if _, err = db.Exec(createDailyLogsTable); err != nil {
		return fmt.Errorf("failed to create daily_logs table: %v", err)
	}

	// Create index for faster queries
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_daily_logs_date ON daily_logs(date);
	CREATE INDEX IF NOT EXISTS idx_daily_logs_status ON daily_logs(status);
	CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
	`

	if _, err = db.Exec(createIndexes); err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// CreateUser inserts a new user into the database
func CreateUser(phoneNumber, name, task, guiltBuddyPhone, guiltBuddyName string) (*User, error) {
	query := `INSERT INTO users (phone_number, name, task, guilt_buddy_phone, guilt_buddy_name) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, phoneNumber, name, task, guiltBuddyPhone, guiltBuddyName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	user := &User{
		ID:              id,
		PhoneNumber:     phoneNumber,
		Name:            name,
		Task:            task,
		GuiltBuddyPhone: guiltBuddyPhone,
		GuiltBuddyName:  guiltBuddyName,
		CreatedAt:       time.Now(),
		Active:          true,
	}

	return user, nil
}

// GetUserByPhone retrieves a user by phone number
func GetUserByPhone(phoneNumber string) (*User, error) {
	query := `SELECT id, phone_number, name, task, guilt_buddy_phone, guilt_buddy_name, created_at, active FROM users WHERE phone_number = ?`
	user := &User{}
	err := db.QueryRow(query, phoneNumber).Scan(
		&user.ID, &user.PhoneNumber, &user.Name, &user.Task, &user.GuiltBuddyPhone, &user.GuiltBuddyName, &user.CreatedAt, &user.Active,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(userID int64) (*User, error) {
	query := `SELECT id, phone_number, name, task, guilt_buddy_phone, guilt_buddy_name, created_at, active FROM users WHERE id = ?`
	user := &User{}
	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.PhoneNumber, &user.Name, &user.Task, &user.GuiltBuddyPhone, &user.GuiltBuddyName, &user.CreatedAt, &user.Active,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetAllActiveUsers retrieves all active users
func GetAllActiveUsers() ([]User, error) {
	query := `SELECT id, phone_number, name, task, guilt_buddy_phone, guilt_buddy_name, created_at, active FROM users WHERE active = 1`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.PhoneNumber, &user.Name, &user.Task, &user.GuiltBuddyPhone, &user.GuiltBuddyName, &user.CreatedAt, &user.Active); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// CreateDailyLog creates a new daily log entry
func CreateDailyLog(userID int64, date string) (*DailyLog, error) {
	query := `INSERT INTO daily_logs (user_id, date, reminder_sent_at, status) VALUES (?, ?, ?, 'pending')`
	result, err := db.Exec(query, userID, date, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create daily log: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	log := &DailyLog{
		ID:             id,
		UserID:         userID,
		Date:           date,
		ReminderSentAt: time.Now(),
		Status:         "pending",
	}

	return log, nil
}

// UpdateDailyLogResponse updates the response for a daily log
func UpdateDailyLogResponse(userID int64, date, response string) error {
	query := `UPDATE daily_logs SET response = ?, responded_at = ?, status = ? WHERE user_id = ? AND date = ?`
	status := "completed"
	if response == "no" || response == "No" || response == "NO" {
		status = "completed"
	}

	_, err := db.Exec(query, response, time.Now(), status, userID, date)
	if err != nil {
		return fmt.Errorf("failed to update daily log response: %v", err)
	}
	return nil
}

// GetPendingLogsForDate retrieves all pending logs for a specific date
func GetPendingLogsForDate(date string) ([]DailyLog, error) {
	query := `SELECT id, user_id, date, reminder_sent_at, response, responded_at, follow_up_sent_at, status
	          FROM daily_logs WHERE date = ? AND status = 'pending'`
	rows, err := db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []DailyLog
	for rows.Next() {
		var log DailyLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.Date, &log.ReminderSentAt,
			&log.Response, &log.RespondedAt, &log.FollowUpSentAt, &log.Status); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// MarkLogAsIgnored marks a daily log as ignored
func MarkLogAsIgnored(userID int64, date string) error {
	query := `UPDATE daily_logs SET status = 'ignored' WHERE user_id = ? AND date = ?`
	_, err := db.Exec(query, userID, date)
	if err != nil {
		return fmt.Errorf("failed to mark log as ignored: %v", err)
	}
	return nil
}

// MarkFollowUpSent marks that a follow-up reminder was sent
func MarkFollowUpSent(userID int64, date string) error {
	query := `UPDATE daily_logs SET follow_up_sent_at = ? WHERE user_id = ? AND date = ?`
	_, err := db.Exec(query, time.Now(), userID, date)
	if err != nil {
		return fmt.Errorf("failed to mark follow-up sent: %v", err)
	}
	return nil
}

// GetUserDailyLogs retrieves all daily logs for a specific user
func GetUserDailyLogs(userID int64, limit int) ([]DailyLog, error) {
	query := `SELECT id, user_id, date, reminder_sent_at, response, responded_at, follow_up_sent_at, status
	          FROM daily_logs WHERE user_id = ? ORDER BY date DESC LIMIT ?`
	rows, err := db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []DailyLog
	for rows.Next() {
		var log DailyLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.Date, &log.ReminderSentAt,
			&log.Response, &log.RespondedAt, &log.FollowUpSentAt, &log.Status); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetDailyLogByUserAndDate retrieves a specific daily log
func GetDailyLogByUserAndDate(userID int64, date string) (*DailyLog, error) {
	query := `SELECT id, user_id, date, reminder_sent_at, response, responded_at, follow_up_sent_at, status
	          FROM daily_logs WHERE user_id = ? AND date = ?`
	log := &DailyLog{}
	err := db.QueryRow(query, userID, date).Scan(
		&log.ID, &log.UserID, &log.Date, &log.ReminderSentAt,
		&log.Response, &log.RespondedAt, &log.FollowUpSentAt, &log.Status,
	)
	if err != nil {
		return nil, err
	}
	return log, nil
}
