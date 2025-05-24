# Thrippy Setup: `slack-oauth`

1. Create the link

   ```shell
   thrippy create-link --template slack-oauth \
           --oauth 'client_id: "..." client_secret: "..."'
   ```

2. Install and authorize the Slack app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```
