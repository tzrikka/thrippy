# Thrippy Setup: `github-user-pat`

1. Create the link

   ```shell
   thrippy create-link --template github-user-pat
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "pat=..."
   ```

## GitHub Enterprise Server

To use GHES instead of the default base URL (`https://github.com`):

Add the following flag to the `set-creds` command in step 2:

```shell
--kv "base_url=https://..."
```

## References

- [Authenticating to the REST API](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api?apiVersion=2022-11-28#authenticating-with-a-personal-access-token)
- [Managing your personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
