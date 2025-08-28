# Slack App Manifest

This directory contains the Slack app manifest files that define the configuration for the Kudos Bot Slack application.

## Files

- `manifest.json` - JSON format manifest
- `manifest.yaml` - YAML format manifest (same configuration)

## Usage

### Creating a Slack App with the Manifest

1. Go to [Slack API: Your Apps](https://api.slack.com/apps)
2. Click "Create New App"
3. Select "From an app manifest"
4. Choose your workspace
5. Copy and paste the content from either `manifest.json` or `manifest.yaml`
6. Review the configuration and click "Create"

### Configuration Required

After creating the app with the manifest, you'll need to:

1. **Update URLs**: Replace `https://your-domain.com` with your actual domain in:
   - OAuth redirect URL: `https://your-domain.com/auth/slack/callback`
   - Slash command URL: `https://your-domain.com/kudos`
   - Event subscriptions URL: `https://your-domain.com/slack/events` (if used)

2. **Get App Credentials**: From the Slack app settings, copy:
   - Client ID
   - Client Secret  
   - Signing Secret

3. **Set Environment Variables**:
   ```bash
   export SLACK_CLIENT_ID="your_client_id"
   export SLACK_CLIENT_SECRET="your_client_secret"
   export SLACK_SIGNING_SECRET="your_signing_secret"
   export REDIRECT_URI="https://your-domain.com/auth/slack/callback"
   ```

### App Features Defined

The manifest configures the following features:

- **Bot User**: A bot user named "Kudos Bot" for posting messages
- **Slash Command**: `/kudos` command for giving kudos to teammates
- **OAuth Scopes**:
  - `commands` - To receive slash commands
  - `chat:write` - To post messages in channels
  - `users:read` - To resolve user IDs to usernames

### Installation Flow

Once the app is configured:

1. Users visit `/auth/slack` to start the OAuth installation
2. They authorize the app in their workspace
3. The app stores the installation with proper tokens
4. Users can then use `/kudos @username description` in any channel

## Customization

You can customize the manifest by:

- Changing the app name and description in `display_information`
- Modifying the slash command URL to match your deployment
- Adding additional OAuth scopes if needed
- Enabling additional features like event subscriptions or interactivity

## Security Notes

- The manifest includes standard security settings
- Token rotation is disabled by default (can be enabled if needed)
- Socket mode is disabled (using HTTP endpoints instead)
- Event subscriptions are configured but no events are enabled by default