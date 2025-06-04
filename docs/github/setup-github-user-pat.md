# Thrippy Setup: `github-user-pat`

1. Create the link

   ```shell
   thrippy create-link --template github-user-pat
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "pat=..."
   ```

## Optional: Webhook Events

If you also defined a personal [webhook](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks) to receive asynchronous event notifications from an organization or repository, and you want to treat the PAT and the webhook as a single unified entity, add the following flag to the `set-creds` command in step 2:

```shell
--kv "webhook_secret=..."
```

## Optional: GitHub Enterprise Server

To use GHES instead of the default base URL (`https://github.com`), add the following flag to the `set-creds` command in step 2:

```shell
--kv "base_url=https://..."
```

## References

- [Authenticating to the REST API](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api?apiVersion=2022-11-28#authenticating-with-a-personal-access-token)
- [Managing your personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
- [Creating webhooks](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks)
