# Thrippy Link Setup: `slack-oauth`

## Slack App Minimum Settings

1. [Create a new Slack app](https://api.slack.com/apps?new_app=1)

   - How to configure the app: "From sractch"
   - App name
   - Pick a Slack workspace
   - Click the "Create App" button

2. Left Panel > Features > OAuth & Permissions

   - "Redirect URLs" section
     - Click the "Add New Redirect URL" button
     - Redirect URL: `https://ADDRESS/callback`\
       (`ADDRESS` is Thrippy's public address for HTTP webhooks - see the `server` command's `-w` flag)
     - Click the "Add" button
     - Click the "Save URLs" button
   - "Scopes" section
     - Bot Token Scopes > click the "Add an OAuth Scope" button
     - Select [`users:read`](https://docs.slack.dev/reference/scopes/users.read)

## Slack App Details to Copy

- Left Panel > Settings > Basic Information
  - Client ID
  - Client secret (click the "Show" button)
  - Signing secret (click the "Show" button)
- Left Panel > Settings > Install App
  - Bot User OAuth Token (click the "Copy" button)

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template slack-oauth \
           --client-id "..." --client-secret "..." \
           [ --scopes "xxx,yyy,..." [ --scopes "zzz" ] ]
   ```

2. Install and authorize the Slack app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```

## References

- [Installing with OAuth](https://docs.slack.dev/authentication/installing-with-oauth)
