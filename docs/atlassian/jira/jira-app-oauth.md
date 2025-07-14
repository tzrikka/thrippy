# Thrippy Link Setup: `jira-app-oauth`

## Jira App Minimum Settings

1. [Atlassian Developer Console > Create a new OAuth 2.0 (3LO) integration](https://developer.atlassian.com/console/myapps/create-3lo-app)

2. Left Sidebar > Permissions

   - User identity API
     - View active user profile (`read:me`)

3. Left Sidebar > Authorization

   - Callback URL: `https://ADDRESS/callback`\
     (`ADDRESS` is Thrippy's [public address for HTTP webhooks](/docs/http_tunnel.md))

## Jira App Details to Copy

- Left Sidebar > Settings
  - Client ID
  - Secret

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template jira-app-oauth \
           --client-id "..." --client-secret "..." \
           [ --scopes "xxx,yyy,..." [ --scopes "zzz" ] ]
   ```

2. Install and authorize the Jira app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```

## References

- [Jira Cloud OAuth 2.0 (3LO) apps](https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/)
