# Thrippy Setup: `slack-bot-token`

1. Create the link

   ```shell
   thrippy create-link --template slack-bot-token
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "bot_token=..."
   ```

## Optional: Events, Interactivity, Slash Commands

If you configured the app to send asynchronous event notifications to a webhook, add the following flag to the `set-creds` command in step 2:

```shell
--kv "signing_secret=..."
```
