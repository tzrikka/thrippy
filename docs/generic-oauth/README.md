# Thrippy Setup: `generic-oauth`

1. Create the link

   ```shell
   thrippy create-link --template generic-oauth \
           --oauth 'auth_url: "..." token_url: "..." \
           client_id: "..." client_secret: "..." scopes: "..."'
   ```

2. Authorize the GitHub app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```
