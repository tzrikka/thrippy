# Thrippy Setup: `slack-oauth-gov`

1. Create the link

   ```shell
   thrippy create-link --template slack-oauth-gov \
           --client-id "..." --client-secret "..." \
           [ --scopes "xxx,yyy,..." [ --scopes "zzz" ] ]
   ```

2. Install and authorize the Slack app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```
