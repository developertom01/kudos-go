# Kudos Go - Multi-Platform OAuth Implementation

This application now includes complete OAuth flows for both Slack and Google Chat platforms, enabling proper app installation and authentication.

## Supported Platforms

### Slack
- **Native @mention support** with Slack's `<@U1234567890>` format
- **Legacy @username support** for backward compatibility  
- **OAuth 2.0 flow** for secure workspace installation
- **Per-installation token management** for multi-workspace support

### Google Chat
- **Native @mention support** with Google Chat's `<users/USER_ID>` format
- **Legacy @username support** for backward compatibility
- **OAuth 2.0 flow** for secure workspace installation
- **Webhook-based messaging** for real-time interaction

## New Features Added

### 1. Multi-Platform OAuth Configuration

#### Slack OAuth
- `SLACK_CLIENT_ID` - Your Slack app's client ID
- `SLACK_CLIENT_SECRET` - Your Slack app's client secret  
- `SLACK_SIGNING_SECRET` - Your Slack app's signing secret for request verification
- `REDIRECT_URI` - OAuth redirect URI (e.g., `https://yourdomain.com/auth/slack/callback`)

#### Google Chat OAuth  
- `GOOGLE_CLIENT_ID` - Your Google Cloud project's OAuth client ID
- `GOOGLE_CLIENT_SECRET` - Your Google Cloud project's OAuth client secret
- `GOOGLE_PROJECT_ID` - Your Google Cloud project ID
- `GOOGLE_REDIRECT_URI` - OAuth redirect URI (e.g., `https://yourdomain.com/auth/googlechat/callback`)
- `GOOGLE_CHAT_WEBHOOK_TOKEN` - Webhook verification token

### 2. OAuth Endpoints

#### Slack
- **`/auth/slack` (GET)** - Initiates the Slack OAuth flow
- **`/auth/slack/callback` (GET)** - Handles the OAuth callback from Slack

#### Google Chat
- **`/auth/googlechat` (GET)** - Initiates the Google Chat OAuth flow
- **`/auth/googlechat/callback` (GET)** - Handles the OAuth callback from Google
- **`/googlechat/webhook` (POST)** - Webhook endpoint for Google Chat events

### 3. Enhanced Database Models

The `Installation` model supports both platforms:
- `Platform` - Identifies the platform ("slack" or "googlechat")
- `AccessToken` - OAuth access token for the workspace
- `BotUserOAuthToken` - Bot user OAuth token for API calls (or refresh token for Google Chat)
- `TeamID` - Platform-specific team/workspace/space ID
- `TeamName` - Platform-specific team/workspace/space name

### 4. Authentication Middleware

Platform-specific middleware that:
- Verifies platform request signatures (configurable)
- Ensures requests come from authorized workspaces
- Protects command endpoints

### 5. Per-Installation Token Management

The command handlers now:
- Retrieve the correct installation for the requesting team/space
- Use the installation-specific tokens for API calls
- Ensure proper isolation between different workspace installations

## Installation Flows

### Slack Installation
1. User visits `/auth/slack` to start installation
2. User is redirected to Slack for authorization
3. Slack redirects back to `/auth/slack/callback` with authorization code
4. App exchanges code for tokens and stores installation
5. User sees success page confirming installation

### Google Chat Installation  
1. User visits `/auth/googlechat` to start installation
2. User is redirected to Google for authorization
3. Google redirects back to `/auth/googlechat/callback` with authorization code
4. App exchanges code for tokens and stores installation
5. User sees success page confirming installation
6. Configure webhook URL in Google Chat API console to point to `/googlechat/webhook`

## Environment Setup

```bash
# Slack Configuration
export SLACK_CLIENT_ID="your_client_id"
export SLACK_CLIENT_SECRET="your_client_secret"  
export SLACK_SIGNING_SECRET="your_signing_secret"
export REDIRECT_URI="https://yourdomain.com/auth/slack/callback"

# Google Chat Configuration
export GOOGLE_CLIENT_ID="your_google_client_id.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET="your_google_client_secret"
export GOOGLE_PROJECT_ID="your_google_project_id"
export GOOGLE_REDIRECT_URI="https://yourdomain.com/auth/googlechat/callback"
export GOOGLE_CHAT_WEBHOOK_TOKEN="your_webhook_verification_token"

# Application Configuration
export KUDOS_SLASH_COMMAND="/kudos"
export PORT=":8080"
```

## Security Features

- **Request signature verification** (platform-specific)
- **State parameter validation** (TODO: implement secure state storage)
- **Per-installation token isolation**
- **Secure token storage in database**
- **Platform-specific authentication**

## Usage

After installation, users can use the `/kudos` command with native @mention functionality on both platforms:

### Slack @mention format (recommended):
```
/kudos <@username> Great work on the project!
```

### Google Chat @mention format (recommended):
```
/kudos <users/USER_ID> Excellent debugging skills!
```

### Legacy @username format (both platforms):
```
/kudos @username Great work on the project!
```

### Features:
- **Native platform @mentions**: Use each platform's autocomplete @mention feature for easy user selection
- **Multi-word descriptions**: Full support for detailed kudos messages
- **User ID resolution**: Automatically resolves platform user IDs to usernames
- **Rich responses**: Responses include platform-appropriate @mentions and emoji for better visibility
- **Cross-platform isolation**: Each platform installation is completely isolated

## Running Platform-Specific Servers

### Slack Server
```bash
cd slack
go run .
```

### Google Chat Server  
```bash
cd googlechat
go run .
```

Both servers can run simultaneously on different ports by configuring different PORT environment variables.

The app will now use the proper OAuth tokens for each workspace installation and display rich messages in the appropriate platform format.