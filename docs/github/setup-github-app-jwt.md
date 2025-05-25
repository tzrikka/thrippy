# Thrippy Setup: `github-app-jwt`

1. Create the link

   ```shell
   thrippy create-link --template github-app-jwt \
           --client-id "..." --param "app_name=..."
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "client_id=..." --kv "private_key=..."
   ```

   The `private_key` value can be:

   - The path of the PEM file (with a `@` prefix): `"private_key=@/path/to/file.pem"`

   - The contents of the PEM file:\
     `"private_key=-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"`

3. Install the GitHub app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```

## GitHub Enterprise Server

To use GHES instead of the default base URL (`https://github.com`):

Add the following flag to the `create-link` command in step 1:

```shell
--param "base_url=http://..."
```
