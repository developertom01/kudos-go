# Google Chat App Configuration Guide

This guide provides step-by-step instructions for setting up a production-ready Google Chat app for the Kudos Bot.

## Prerequisites

- Google Cloud Project with billing enabled
- Google Workspace domain (for domain-wide installation)
- Domain with HTTPS capability for webhook endpoints

## Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Select a project" → "New Project"
3. Enter project name: `kudos-chat-bot`
4. Select your organization (if applicable)
5. Click "Create"

## Step 2: Enable APIs

1. In the Google Cloud Console, go to "APIs & Services" → "Library"
2. Search for and enable:
   - **Google Chat API**
   - **Cloud Resource Manager API** (if using service accounts)

## Step 3: Configure Google Chat API

1. Go to "APIs & Services" → "Google Chat API" → "Configuration"
2. Fill in the app details:

### App Information
```
App name: Kudos Bot
Avatar URL: https://your-domain.com/assets/kudos-bot-icon.png (optional)
Description: Give kudos and recognition to your teammates
```

### Interactive Features
```
✅ Enable interactive features
Status: LIVE
```

### Functionality
```
✅ Receive 1:1 messages
✅ Join spaces and group conversations
✅ Log interactions to Google Cloud Logging (optional)
```

### Connection Settings
```
Bot type: HTTP endpoint
Bot URL: https://your-domain.com/googlechat/webhook
```

### Slash Commands
```
Command: /kudos
URL: https://your-domain.com/googlechat/webhook  
Description: Give kudos to a teammate
```

### Permissions
For domain-wide deployment:
```
✅ Make this Google Chat app available to specific people and groups in [your-domain.com]
```

For limited testing:
```
✅ Make this Google Chat app available to specific people and groups in your domain
Add specific users/groups for testing
```

## Step 4: Set up OAuth 2.0

1. Go to "APIs & Services" → "Credentials"
2. Click "Create Credentials" → "OAuth 2.0 Client IDs"
3. Configure the OAuth client:

```
Application type: Web application
Name: Kudos Bot OAuth Client

Authorized JavaScript origins:
- https://your-domain.com

Authorized redirect URIs:
- https://your-domain.com/auth/googlechat/callback
```

4. Click "Create" and save the Client ID and Client Secret

## Step 5: Configure Webhook Token

For webhook security, generate a secure token:

```bash
# Generate a secure random token
openssl rand -base64 32
```

Save this token - you'll need it for both:
- Environment variable: `GOOGLE_CHAT_WEBHOOK_TOKEN`
- Google Chat app configuration (if using token verification)

## Step 6: Environment Configuration

Create your production environment configuration:

```bash
# Google Chat OAuth Configuration
GOOGLE_CLIENT_ID="123456789.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="GOCSPX-your_client_secret_here"
GOOGLE_PROJECT_ID="kudos-chat-bot"
GOOGLE_REDIRECT_URI="https://your-domain.com/auth/googlechat/callback"

# Webhook Security
GOOGLE_CHAT_WEBHOOK_TOKEN="your_secure_random_token_here"

# Application Configuration
PORT=":8080"
GIN_MODE="release"

# Database Configuration
DATABASE_URL="postgres://username:password@db-host:5432/kudos_prod?sslmode=require"
```

## Step 7: Deploy Application

### Using Docker

1. **Build and deploy**:
```bash
docker build -t kudos-googlechat .
docker run -d \
  --name kudos-googlechat \
  -p 8080:8080 \
  --env-file .env.production \
  kudos-googlechat
```

### Using Cloud Run

1. **Build and push to Container Registry**:
```bash
gcloud builds submit --tag gcr.io/your-project-id/kudos-googlechat
```

2. **Deploy to Cloud Run**:
```bash
gcloud run deploy kudos-googlechat \
  --image gcr.io/your-project-id/kudos-googlechat \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --set-env-vars GOOGLE_CLIENT_ID=your_client_id \
  --set-env-vars GOOGLE_CLIENT_SECRET=your_client_secret \
  --set-env-vars GOOGLE_PROJECT_ID=your_project_id \
  --set-env-vars GOOGLE_CHAT_WEBHOOK_TOKEN=your_webhook_token
```

## Step 8: Update Google Chat Configuration

After deployment, update the Google Chat app configuration with your actual URLs:

1. Go to Google Chat API → Configuration
2. Update **Bot URL**: `https://your-actual-domain.com/googlechat/webhook`
3. Update **Slash command URL**: `https://your-actual-domain.com/googlechat/webhook`
4. Save the configuration

## Step 9: Test Installation

1. **Health check**: Visit `https://your-domain.com/health`
2. **OAuth flow**: Visit `https://your-domain.com/auth/googlechat`
3. **Test in Google Chat**: 
   - Add the bot to a space
   - Try: `/kudos @someone Great work!`

## Step 10: Domain-Wide Deployment

For organization-wide deployment:

### Method 1: Google Workspace Admin Console

1. Go to [Google Admin Console](https://admin.google.com/)
2. Navigate to "Apps" → "Google Workspace" → "Google Chat"
3. Go to "Manage Apps"
4. Find your app or add it using the app ID
5. Set availability for your organization

### Method 2: App Marketplace (Optional)

For wider distribution, consider publishing to Google Workspace Marketplace:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to "APIs & Services" → "Google Workspace Marketplace SDK"
3. Follow the publishing guidelines

## Security Best Practices

### Production Security Checklist

- [ ] **HTTPS Only** - All endpoints use HTTPS
- [ ] **Token Verification** - Webhook token properly configured
- [ ] **OAuth Scopes** - Minimal required scopes only
- [ ] **State Validation** - CSRF protection enabled
- [ ] **Rate Limiting** - Request rate limiting configured
- [ ] **Error Handling** - No sensitive data in error messages
- [ ] **Logging** - Comprehensive security event logging
- [ ] **Database Security** - Connection encryption and credentials protection

### Monitoring and Maintenance

- **Health Monitoring**: Monitor `/health` endpoint
- **Log Analysis**: Regular review of application logs
- **Token Rotation**: Plan for OAuth token refresh
- **Updates**: Keep dependencies updated
- **Backup**: Regular database backups

## Troubleshooting Production Issues

### Common Deployment Issues

**Google Chat app not receiving events**:
- Verify webhook URL is publicly accessible
- Check HTTPS certificate validity
- Confirm webhook token matches
- Review Google Chat app permissions

**OAuth installation fails**:
- Check redirect URI matches exactly
- Verify Client ID and Secret are correct
- Ensure OAuth consent screen is configured
- Check that Google Chat API is enabled

**Database connection issues**:
- Verify DATABASE_URL format
- Check network connectivity
- Confirm database credentials
- Review connection pool settings

## Support and Resources

- [Google Chat API Documentation](https://developers.google.com/chat)
- [Google Cloud Console](https://console.cloud.google.com/)
- [OAuth 2.0 Best Practices](https://tools.ietf.org/html/rfc6749)
- [Google Workspace Admin Help](https://support.google.com/a/answer/7651360)

## Configuration Templates

### Development Environment
```bash
GOOGLE_CLIENT_ID="dev-client-id.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="dev-client-secret"
GOOGLE_PROJECT_ID="kudos-dev"
GOOGLE_REDIRECT_URI="http://localhost:8081/auth/googlechat/callback"
GOOGLE_CHAT_WEBHOOK_TOKEN="dev-webhook-token"
PORT=":8081"
GIN_MODE="debug"
DATABASE_URL="postgres://dev:dev@localhost:5432/kudos_dev?sslmode=disable"
```

### Production Environment
```bash
GOOGLE_CLIENT_ID="prod-client-id.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="prod-client-secret"
GOOGLE_PROJECT_ID="kudos-prod"
GOOGLE_REDIRECT_URI="https://kudos.yourdomain.com/auth/googlechat/callback"
GOOGLE_CHAT_WEBHOOK_TOKEN="secure-random-token-256-bits"
PORT=":8080"
GIN_MODE="release"
DATABASE_URL="postgres://kudos:secure-password@db.internal:5432/kudos_prod?sslmode=require"
```