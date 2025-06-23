# Thrippy Link Setup: `google-user-oauth`

## Google Cloud OAuth

Minimum settings and details to copy: [see this page](./gcp-oauth.md).

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template google-user-oauth \
           --client-id "..." --client-secret "..." \
           [ --scopes "xxx,yyy,..." [ --scopes "zzz" ] ]
   ```

2. Authorize the GitHub app (interactively in a browser) ...

   ```shell
   thrippy start-oauth <link ID>
   ```
