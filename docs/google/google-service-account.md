# Thrippy Link Setup: `google-service-account`

## Google Cloud OAuth

Minimum settings and details to copy: [see this page](./gcp-oauth.md).

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template google-service-account
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "key=..."
   ```

   The `key` value can be:

   - The path of the JSON file (with a `@` prefix): `"private_key=@/path/to/file.json"`

   - The contents of the JSON file: `"key={ "type": "service_account", ... }"`
