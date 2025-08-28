# Kudos Go - Slack OAuth Implementation

This application now includes a complete Slack OAuth flow for proper app installation and authentication.

## New Features Added

### 1. OAuth Configuration
Added environment variables for Slack OAuth:
- `SLACK_CLIENT_ID` - Your Slack app's client ID
- `SLACK_CLIENT_SECRET` - Your Slack app's client secret  
- `SLACK_SIGNING_SECRET` - Your Slack app's signing secret for request verification
- `REDIRECT_URI` - OAuth redirect URI (e.g., `https://yourdomain.com/auth/slack/callback`)

### 2. OAuth Endpoints

#### `/auth/slack` (GET)
Initiates the Slack OAuth flow. Redirects users to Slack's authorization page.

#### `/auth/slack/callback` (GET)  
Handles the OAuth callback from Slack. Exchanges authorization code for access tokens and stores the installation.

### 3. Enhanced Database Models

The `Installation` model now includes:
- `AccessToken` - OAuth access token for the workspace
- `BotUserOAuthToken` - Bot user OAuth token for API calls
- `TeamID` - Slack team/workspace ID
- `TeamName` - Slack team/workspace name

### 4. Authentication Middleware

Added middleware that:
- Verifies Slack request signatures (configurable)
- Ensures requests come from authorized Slack workspaces
- Protects slash command endpoints

### 5. Per-Installation Token Management

The slash command handler now:
- Retrieves the correct installation for the requesting team
- Uses the installation-specific bot token for API calls
- Ensures proper isolation between different workspace installations

## Installation Flow

1. User visits `/auth/slack` to start installation
2. User is redirected to Slack for authorization
3. Slack redirects back to `/auth/slack/callback` with authorization code
4. App exchanges code for tokens and stores installation
5. User sees success page confirming installation

## Environment Setup

```bash
export SLACK_CLIENT_ID="your_client_id"
export SLACK_CLIENT_SECRET="your_client_secret"  
export SLACK_SIGNING_SECRET="your_signing_secret"
export REDIRECT_URI="https://yourdomain.com/auth/slack/callback"
export KUDOS_SLASH_COMMAND="/kudos"
export PORT=":8080"
```

## Security Features

- Request signature verification (Slack signing secret)
- State parameter validation (TODO: implement secure state storage)
- Per-installation token isolation
- Secure token storage in database

## Usage

After installation, users can use the `/kudos` command:
```
/kudos @username Great work on the project!
```

The app will now use the proper OAuth tokens for each workspace installation.