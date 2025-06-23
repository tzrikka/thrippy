# Thrippy Link Setup: `slack-bot-token`

## Slack App Minimum Settings

1. [Create a new Slack app](https://api.slack.com/apps?new_app=1)

   - How to configure the app: "From sractch"
   - App name
   - Pick a Slack workspace
   - Click the "Create App" button

2. Left Panel > Features > OAuth & Permissions

   - "Scopes" section
     - Bot Token Scopes > click the "Add an OAuth Scope" button
     - Select (at least) [`users:read`](https://docs.slack.dev/reference/scopes/users.read)

3. Left Panel > Settings > Install App

   - Click the "Install to Workspace" button
   - Click the "Allow" button

## Slack App Details to Copy

- Left Panel > Settings > Basic Information
  - Signing secret
- Left Panel > Settings > Install App
  - Bot User OAuth Token (click the "Copy" button)

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template slack-bot-token
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "bot_token=..." --kv "signing_secret=..."
   ```
