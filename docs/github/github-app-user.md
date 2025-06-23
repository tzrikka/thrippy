# Thrippy Link Setup: `github-app-user`

## GitHub App

Minimum settings and details to copy: [see this page](./app-settings.md).

## Thrippy Link Setup

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

## Optional: GitHub Enterprise Server

To use GHES instead of the default base URL (`https://github.com`), run the following command between steps 1 and 2:

```shell
set-creds <link ID> --kv "base_url=https://..."
```
