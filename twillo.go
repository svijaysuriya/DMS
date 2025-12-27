package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func sendTwilioWhatsApp(to, message string) error {
	// Load credentials from environment variables
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromWhatsApp := os.Getenv("TWILIO_WHATSAPP_NUMBER")

	// Validate that required environment variables are set
	if accountSid == "" || authToken == "" || fromWhatsApp == "" {
		log.Println("ERROR: Twilio credentials not set in environment variables")
		return fmt.Errorf("missing Twilio credentials in environment variables")
	}

	apiURL := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"

	data := url.Values{}
	data.Set("To", "whatsapp:"+to) // e.g., "whatsapp:+919999999999"
	data.Set("From", fromWhatsApp)
	data.Set("Body", message)

	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("Message sent successfully!")
		return nil
	} else {
		return fmt.Errorf("failed to send message, status: %s", resp.Status)
	}
}
