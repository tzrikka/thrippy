# GitHub Apps

## Apps Acting on Their Own Behalf

App settings:

- Post-installation setup URL: `https://ADDRESS/callback`\
  (`ADDRESS` is Trippy's public address for HTTP webhooks - see the `server` command's `-w` flag)
- Post-installation redirect on update: yes
- Generate a private key

Details to copy:

- App name (the URL slug, not the display name)
- Client ID
- Downloaded private key PEM file

Trippy setup:

1. `create-link --template github-app-jwt --oauth 'client_id: "..." params: { key: "app_name" value: "..."}'`
2. `set-creds <link ID> --kv "client_id=..." --kv "private_key=..."`
3. `start-oauth <link ID>`

**TODO:** The private key value can be the contents of the PEM file, or its path (prefixed with `@`), or `-` to read it from stdin.

## Apps Acting on Behalf of Users

App settings:

- Generate a client secret
- Callback URL: `https://ADDRESS/callback`\
  (`ADDRESS` is Trippy's public address for HTTP webhooks - see the `server` command's `-w` flag)

Details to copy:

- Client ID
- Client secret

Trippy setup:

1. `create-link --template github-app-user --oauth 'client_id: "..." client_secret: "..."'`
2. `start-oauth <link ID>`

## GitHub Enterprise Server (GHES)

The default base URL is `https://github.com`.

To use a different GitHub Enterprise Server (GHES):

- Apps acting on their own behalf (`github-app-jwt` template):\
  append the following to the `--oauth` flag in the `create-link` command:

  ```
  params: { key: "base_url" value: "..." }
  ```

- Apps acting on behalf of users (`github-app-user` template):\
  run the following command between steps 1 and 2:

  ```
  set-creds <link ID> --kv "xxx=..."
  ```

## References

- [About creating GitHub Apps](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps)
- [Registering a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app)
- [Managing private keys for GitHub Apps](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/managing-private-keys-for-github-apps)
