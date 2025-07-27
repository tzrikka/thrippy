# Thrippy Link Setup: `bitbucket-app-oauth`

## Bitbucket App Minimum Settings

1. [Bitbucket Workspaces](https://bitbucket.org/account/workspaces/) > Manage

2. Left Sidebar > Apps and Features > OAuth Consumers

3. Click the "Add consume" button

   - Callback URL: `https://ADDRESS/callback`\
     (`ADDRESS` is Thrippy's [public address for HTTP webhooks](/docs/http_tunnel.md))
   - Permissions
     - Account: Read
     - Webhooks: Read and write
   - Click the "Save" button

## Bitbucket App Details to Copy

- Click the consumer name to see its details
  - Key (client ID)
  - Secret

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template bitbucket-app-oauth \
           --client-id "..." --client-secret "..." \
           [ --scopes "xxx,yyy,..." [ --scopes "zzz" ] ]
   ```

2. Install and authorize the Bitbucket app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```

## References

- [Bitbucket Cloud OAuth 2.0](https://developer.atlassian.com/cloud/bitbucket/oauth-2/)
- [Use OAuth on Bitbucket Cloud](https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/)
- [Bitbucket OAuth 2.0 scopes](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#bitbucket-oauth-2-0-scopes)
