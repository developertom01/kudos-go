# Google Chat Kudos Bot

A production-ready Google Chat bot for giving kudos to team members with complete OAuth 2.0 integration, native @mention support, and comprehensive security features.

## Features

- **OAuth 2.0 Authentication** - Secure app installation flow for Google Chat workspaces
- **Native @mentions** - Support for Google Chat's `<users/USER_ID>` format and legacy `@username`
- **Slash Commands** - `/kudos` command for giving recognition to teammates
- **Multi-word Descriptions** - Full support for detailed kudos messages
- **Production Security** - Request verification, state parameter validation, and rate limiting
- **Comprehensive Logging** - Detailed logging for monitoring and debugging
- **Health Monitoring** - Health check endpoint with detailed status
- **Error Handling** - Graceful error handling with user-friendly messages

## Quick Start

### 1. Create a Google Chat App

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Chat API
4. Go to [Google Chat API Configuration](https://console.cloud.google.com/apis/api/chat.googleapis.com/hangouts-chat)
5. Click "Configuration" tab
6. Configure your app:
   - **App name**: Kudos Bot
   - **Avatar URL**: (optional)
   - **Description**: Give kudos to your teammates
   - **Interactive features**: Enable
   - **Functionality**: 
     - ✅ Receive 1:1 messages
     - ✅ Join spaces and group conversations
   - **Connection settings**: HTTP endpoint
   - **Bot URL**: `https://your-domain.com/googlechat/webhook`
   - **Slash commands**: 
     - Command: `/kudos`
     - URL: `https://your-domain.com/googlechat/webhook`
     - Description: Give kudos to a teammate
   - **Permissions**: Specific people and groups in your domain

### 2. Set up OAuth Credentials

1. In Google Cloud Console, go to "Credentials"
2. Click "Create Credentials" → "OAuth 2.0 Client IDs"
3. Application type: "Web application"
4. Add authorized redirect URI: `https://your-domain.com/auth/googlechat/callback`
5. Copy the Client ID and Client Secret

### 3. Configure Environment Variables

Create a `.env` file or set environment variables:

```bash
# Required - Google Chat OAuth
GOOGLE_CLIENT_ID="your_client_id.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="your_client_secret"
GOOGLE_PROJECT_ID="your_google_project_id"
GOOGLE_REDIRECT_URI="https://your-domain.com/auth/googlechat/callback"

# Required - Webhook Security
GOOGLE_CHAT_WEBHOOK_TOKEN="your_webhook_verification_token"

# Optional - Application Configuration
PORT=":8081"
KUDOS_SLASH_COMMAND="/kudos"

# Database (PostgreSQL connection string)
DATABASE_URL="postgres://username:password@localhost/kudos_db?sslmode=disable"
```

### 4. Deploy and Install

1. **Deploy the application**:
   ```bash
   cd googlechat
   go run .
   ```

2. **Install the app**: Visit `https://your-domain.com/auth/googlechat`

3. **Start using**: Use `/kudos @user description` in any Google Chat space

## Usage Examples

```
/kudos <users/123456789> Great work on the deployment! 🚀
/kudos @john Excellent debugging skills!
/kudos <users/987654321> Amazing presentation today! Thanks for the insights.
/kudos @sarah Outstanding code review, caught several important issues.
```

## Configuration Reference

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GOOGLE_CLIENT_ID` | OAuth 2.0 Client ID from Google Cloud Console | `123456.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | OAuth 2.0 Client Secret | `GOCSPX-abc123...` |
| `GOOGLE_PROJECT_ID` | Google Cloud Project ID | `my-kudos-app` |
| `GOOGLE_CHAT_WEBHOOK_TOKEN` | Webhook verification token | `secure_random_token_123` |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `:8081` | Server port |
| `GOOGLE_REDIRECT_URI` | `http://localhost:8081/auth/googlechat/callback` | OAuth callback URL |
| `KUDOS_SLASH_COMMAND` | `/kudos` | Slash command name |
| `DATABASE_URL` | - | PostgreSQL connection string |

## Security Features

### OAuth 2.0 Security
- **State Parameter Validation** - CSRF protection with cryptographically secure state parameters
- **Token Storage** - Secure storage of access and refresh tokens
- **Scope Limitation** - Minimal required scopes: `chat.bot` and `chat.messages`

### Request Verification
- **Webhook Token Verification** - All webhook requests verified using bearer token
- **Rate Limiting** - Built-in rate limiting (30 requests/minute per IP)
- **Input Validation** - Comprehensive validation of all incoming requests

### Error Handling
- **Graceful Degradation** - Continues operating with limited functionality if database unavailable
- **Detailed Logging** - Comprehensive logging for security monitoring
- **User-Friendly Errors** - Clear error messages without exposing sensitive information

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/googlechat` | GET | Start OAuth installation flow |
| `/auth/googlechat/callback` | GET | OAuth callback handler |
| `/googlechat/webhook` | POST | Google Chat webhook for slash commands |
| `/health` | GET | Health check with detailed status |

## Health Monitoring

The `/health` endpoint provides detailed status information:

```json
{
  "status": "ok",
  "service": "Google Chat Kudos Bot",
  "version": "1.0.0",
  "database": "connected",
  "configuration": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

Status can be:
- `ok` - All systems operational
- `degraded` - Operating with limited functionality
- Returns HTTP 503 if database or configuration issues

## Development

### Running Locally

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Set up environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run the application**:
   ```bash
   go run .
   ```

4. **Use ngrok for webhook testing**:
   ```bash
   ngrok http 8081
   # Update GOOGLE_REDIRECT_URI and webhook URL in Google Chat config
   ```

### Testing

Run the test suite:
```bash
go test ./...
```

Run specific tests:
```bash
go test -v ./...                    # Verbose output
go test -run TestGoogleChatAuth     # Specific test
```

## Deployment

### Docker Deployment

1. **Build the image**:
   ```bash
   docker build -t googlechat-kudos .
   ```

2. **Run the container**:
   ```bash
   docker run -p 8081:8081 \
     -e GOOGLE_CLIENT_ID="your_client_id" \
     -e GOOGLE_CLIENT_SECRET="your_secret" \
     -e GOOGLE_PROJECT_ID="your_project" \
     -e GOOGLE_CHAT_WEBHOOK_TOKEN="your_token" \
     googlechat-kudos
   ```

### Production Deployment

For production deployments:

1. **Use HTTPS** - Google Chat requires HTTPS for webhooks
2. **Set Gin mode**: `export GIN_MODE=release`
3. **Configure logging** - Use structured logging for monitoring
4. **Database setup** - Use PostgreSQL with connection pooling
5. **Health checks** - Monitor the `/health` endpoint
6. **Rate limiting** - Consider using Redis for distributed rate limiting

### Environment-Specific Configuration

**Development**:
```bash
PORT=":8081"
GOOGLE_REDIRECT_URI="http://localhost:8081/auth/googlechat/callback"
GIN_MODE="debug"
```

**Production**:
```bash
PORT=":8080"
GOOGLE_REDIRECT_URI="https://your-domain.com/auth/googlechat/callback"
GIN_MODE="release"
DATABASE_URL="postgres://user:pass@db:5432/kudos_prod?sslmode=require"
```

## Troubleshooting

### Common Issues

**"App not installed" error**:
- Ensure the team has completed OAuth installation via `/auth/googlechat`
- Check database connectivity and installation records

**Webhook not receiving events**:
- Verify webhook URL is accessible from Google's servers (use HTTPS)
- Check `GOOGLE_CHAT_WEBHOOK_TOKEN` matches Google Chat configuration
- Ensure bot is added to the space

**OAuth callback errors**:
- Verify `GOOGLE_REDIRECT_URI` matches Google Cloud Console configuration
- Check `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are correct
- Ensure state parameter is being generated and validated

### Debugging

Enable verbose logging:
```bash
export GIN_MODE=debug
```

Check application logs for:
- OAuth token exchange errors
- Database connection issues  
- Webhook authentication failures
- Rate limiting events

### Support

For issues and questions:
1. Check the logs for detailed error messages
2. Verify all required environment variables are set
3. Test the `/health` endpoint for system status
4. Review Google Chat API documentation for webhook requirements

## License

This project is licensed under the MIT License - see the LICENSE file for details.