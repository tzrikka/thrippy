# Thrippy Setup: `google-user-oauth`

1. Create the link

   ```shell
   thrippy create-link --template google-user-oauth \
           --oauth 'client_id: "..." client_secret: "..."'
   ```

2. Authorize the GitHub app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```
