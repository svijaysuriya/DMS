# Daily Task Reminder System

A Go-based system that sends daily WhatsApp reminders to users about their tasks, tracks responses, and handles non-responses with follow-up messages.

## Features

- **Web Dashboard**: Clean, simple frontend to view and register users
- **User Registration**: Collect phone number, name, and task (via API or web interface)
- **Guilt Buddy System**: Optional accountability partner who gets notified when tasks aren't completed
- **SQLite Database**: Local storage for user data and daily logs
- **Automated Reminders**: Send daily WhatsApp reminders via Twilio
- **Response Tracking**: Receive and store user responses
- **Follow-up System**: Send follow-up reminders to non-responders
- **Ignore Handling**: Mark reminders as ignored if no response by end of day
- **Accountability Notifications**: Notify guilt buddies when users don't complete tasks

## Architecture

### Database Schema

#### Users Table
- `id`: Primary key
- `phone_number`: Unique phone number (with country code)
- `name`: User's name
- `task`: The task to remind about
- `guilt_buddy_phone`: Phone number of accountability buddy (optional)
- `guilt_buddy_name`: Name of accountability buddy (optional)
- `created_at`: Registration timestamp
- `active`: Boolean flag for active users

#### Daily Logs Table
- `id`: Primary key
- `user_id`: Foreign key to users table
- `date`: Date of the reminder (YYYY-MM-DD)
- `reminder_sent_at`: Timestamp when reminder was sent
- `response`: User's response text
- `responded_at`: Timestamp when user responded
- `follow_up_sent_at`: Timestamp when follow-up was sent
- `status`: Status (pending, completed, ignored)

## Web Dashboard

The system includes a clean, simple web interface accessible at `http://localhost:8080/`

### Features

- **User Statistics**: View total users, active users, and users with accountability buddies
- **User List**: Browse all registered users with search functionality
- **User Registration**: Add new users via a modal form
- **Responsive Design**: Works on desktop, tablet, and mobile

### Using the Dashboard

1. Start the server: `go run *.go`
2. Open browser: `http://localhost:8080/`
3. Click **"+ Add User"** to register a new user
4. Use the search box to filter users
5. Click **"Refresh"** to reload data

## API Endpoints

### 1. Dashboard (Frontend)
**GET** `/`

Serves the web dashboard for managing users.

### 2. Register User
**POST** `/register`

Register a new user for daily reminders.

**Request Body:**
```json
{
  "phone_number": "+919999999999",
  "name": "John Doe",
  "task": "Exercise for 30 minutes",
  "guilt_buddy_phone": "+911234567890",
  "guilt_buddy_name": "Jane Doe"
}
```

**Note:** `guilt_buddy_phone` and `guilt_buddy_name` are optional. If provided, the guilt buddy will be notified when the user doesn't complete their task.

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "phone_number": "+919999999999",
    "name": "John Doe",
    "task": "Exercise for 30 minutes",
    "guilt_buddy_phone": "+911234567890",
    "guilt_buddy_name": "Jane Doe",
    "created_at": "2025-12-13T10:00:00Z",
    "active": true
  }
}
```

### 3. Twilio Webhook
**POST** `/webhook/twilio`

Receives incoming WhatsApp messages from Twilio. This endpoint should be configured in your Twilio console.

**Twilio sends form data with:**
- `From`: Sender's WhatsApp number (e.g., "whatsapp:+919999999999")
- `Body`: Message text

**Accepted Responses:**
- "Yes", "Done", "Completed", "Y" → Marks task as completed
- "No", "Not yet", "N" → Acknowledges but marks as not completed
- Other → Asks for clarification

### 4. Get User Info
**GET** `/user?phone=+919999999999`

Retrieve user information by phone number.

### 5. Get User Logs
**GET** `/user/logs?phone=+919999999999`

Retrieve daily logs for a specific user (last 30 days).

### 6. Get All Users
**GET** `/users`

Retrieve all active users.

## Reminder Schedule

### Daily Reminder
- **Time**: 9:00 AM (configurable in `scheduler.go`)
- **Action**: Sends reminder to all active users
- **Message**: "Good morning, {Name}! 🌅\n\nDaily reminder: {Task}\n\nReply 'Yes' or 'Done' when you complete it!"

### Follow-up Reminder
- **Time**: 6:00 PM (configurable in `scheduler.go`)
- **Action**: Sends follow-up to users who haven't responded
- **Message**: "Hi {Name}, just checking in! 👋\n\nHave you completed your task: {Task}?\n\nReply 'Yes' or 'No'"

### Ignore Marking
- **Time**: 11:00 PM
- **Action**: Marks all pending reminders as "ignored"

## Handling Non-Responses

The system has a three-tier approach:

1. **Initial Reminder** (9:00 AM): Sends the daily reminder
2. **Follow-up** (6:00 PM): If no response, sends a follow-up message
3. **Mark as Ignored** (11:00 PM): If still no response, marks the log as "ignored"

### What happens when a user ignores messages?

- The daily log status is set to "ignored"
- **If a guilt buddy is configured**: They receive a notification message
- The system continues to send reminders the next day
- You can query ignored logs to identify users who need attention

## Guilt Buddy (Accountability Partner) Feature

The guilt buddy feature adds social accountability to task completion. When registering, users can optionally provide:
- A phone number of someone they trust (guilt buddy)
- The name of that person

### How It Works

1. **Registration**: User provides guilt buddy's phone number and name
2. **Welcome Notification**: Guilt buddy receives a message informing them they've been added as an accountability partner
3. **Daily Tracking**: System monitors if the user completes their task
4. **Notification**: If the user doesn't respond by 11:00 PM, the guilt buddy receives a message

### Guilt Buddy Messages

**When added as accountability partner:**
```
Hi {GuiltBuddyName}! 👋

{UserName} has added you as their accountability buddy for: {Task}

You'll receive a notification if they don't complete their daily task.
```

**When user doesn't complete task:**
```
Hey {GuiltBuddyName}! 😔

{UserName} didn't complete their task today: {Task}

Maybe check in with them?
```

### Benefits

- **Social Accountability**: Users are more likely to complete tasks knowing someone will be notified
- **Support System**: Guilt buddies can provide encouragement and support
- **Optional**: Users can register without a guilt buddy if they prefer
- **Privacy**: Only notifies when tasks are NOT completed, not when they are

## Setup Instructions

### 1. Install Dependencies

```bash
cd deadManSwitch
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv
```

### 2. Configure Environment Variables

**Important:** All credentials and configuration are now managed through environment variables.

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your actual credentials:
   ```bash
   nano .env  # or use your preferred editor
   ```

3. Required environment variables:
   - `TWILIO_ACCOUNT_SID` - Your Twilio Account SID
   - `TWILIO_AUTH_TOKEN` - Your Twilio Auth Token
   - `TWILIO_WHATSAPP_NUMBER` - Your Twilio WhatsApp number (format: `whatsapp:+14444444444`)
   - `SERVER_PORT` - Server port (default: 8080)
   - `DATABASE_PATH` - Database file path (default: `./reminder_system.db`)
   - `ADMIN_PHONE_NUMBER` - Admin phone for legacy alerts (format: `+919999999999`)
   - `REMINDER_TIME` - Daily reminder time in 24h format (default: `09:00`)
   - `FOLLOWUP_TIME` - Follow-up time in 24h format (default: `18:00`)

### 3. Run the Server

```bash
go run *.go
```

The server will start on the port specified in your `.env` file (default: 8080).

### 4. Configure Twilio Webhook

1. Go to Twilio Console → WhatsApp Sandbox Settings
2. Set the webhook URL to: `https://your-domain.com/webhook/twilio`
3. Use ngrok for local testing: `ngrok http 8080`

## Testing

### Register a User
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+919999999999",
    "name": "Test User",
    "task": "Complete daily standup"
  }'
```

### Get User Info
```bash
curl "http://localhost:8080/user?phone=+919999999999"
```

### Get User Logs
```bash
curl "http://localhost:8080/user/logs?phone=+919999999999"
```

## Database Location

The SQLite database location is configured via the `DATABASE_PATH` environment variable (default: `./reminder_system.db`)

## Customization

### Change Reminder Times

Update the `.env` file:

```env
REMINDER_TIME=09:00  # 9 AM
FOLLOWUP_TIME=18:00  # 6 PM
```

The application will automatically use these times when it starts.

### Customize Messages

Edit the message templates in `scheduler.go` and `handlers.go`.

## Production Considerations

1. **Environment Variables**: ✅ **IMPLEMENTED** - All credentials now use environment variables
2. **HTTPS**: Use HTTPS for the webhook endpoint
3. **Rate Limiting**: Implement rate limiting for API endpoints
4. **Database Backups**: Regular backups of SQLite database
5. **Monitoring**: Add logging and monitoring for failed messages
6. **Error Handling**: Implement retry logic for failed Twilio API calls
7. **User Management**: Add endpoints to deactivate/reactivate users
8. **Secrets Management**: Use cloud secrets manager (AWS Secrets Manager, Google Secret Manager, etc.) in production
9. **Never commit `.env`**: The `.env` file is in `.gitignore` to prevent credential leaks

## Troubleshooting

### Messages not sending
- Check Twilio credentials
- Verify phone numbers include country code (+)
- Check Twilio account balance
- Review server logs for errors

### Webhook not receiving messages
- Verify webhook URL in Twilio console
- Ensure server is publicly accessible
- Check firewall settings
- Test with ngrok for local development

### Database errors
- Ensure write permissions for database file
- Check disk space
- Verify SQLite driver is installed

