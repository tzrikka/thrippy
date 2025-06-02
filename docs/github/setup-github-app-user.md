# Thrippy Setup: `github-app-user`

1. Create the link

   ```shell
   thrippy create-link --template github-app-user \
           --client-id "..." --client-secret "..." \
           [ --scopes "xxx,yyy,..." [ --scopes "zzz" ] ]
   ```

2. Authorize the GitHub app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```

## GitHub Enterprise Server

To use GHES instead of the default base URL (`https://github.com`):

Run the following command between steps 1 and 2:

```
set-creds <link ID> --kv "base_url=https://..."
```
