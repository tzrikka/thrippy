# Thrippy Link Setup: `slack-socket-mode`

## Slack App Minimum Settings

1. [Create a new Slack app](https://api.slack.com/apps?new_app=1)

   - How to configure the app: "From sractch"
   - App name
   - Pick a Slack workspace
   - Click the "Create App" button

2. Left Sidebar > Settings > Socket Mode

   - Enable Socket Mode: on
   - Generate an app-level token to enable Socket Mode
     - Token name
     - Scope: [`connections:write`](https://docs.slack.dev/reference/scopes/connections.write)
     - Click the "Generate" button

3. Left Sidebar > Features > OAuth & Permissions

   - "Scopes" section
     - Bot Token Scopes > click the "Add an OAuth Scope" button
     - Select (at least) [`users:read`](https://docs.slack.dev/reference/scopes/users.read)

4. Left Sidebar > Settings > Install App

   - Click the "Install to Workspace" button
   - Click the "Allow" button

## Slack App Details to Copy

- Left Sidebar > Settings > Basic Information
  - "App-Level Tokens" section
    - Click the token name (click the "Copy" button)
- Left Sidebar > Settings > Install App
  - Bot User OAuth Token (click the "Copy" button)

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template slack-socket-mode
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "app_token=..." --kv "bot_token=..."
   ```

## References

- [Socket Mode](https://docs.slack.dev/apis/events-api/using-socket-mode)
- [Comparing HTTP & Socket Mode](https://docs.slack.dev/apis/events-api/comparing-http-socket-mode)
