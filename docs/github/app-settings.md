# GitHub Apps: Settings & Details

## Apps Acting on Their Own Behalf

App settings:

- Post-installation setup URL: `https://ADDRESS/callback`\
  (`ADDRESS` is Thrippy's public address for HTTP webhooks - see the `server` command's `-w` flag)
- Post-installation redirect on update: yes
- Generate a private key

Details to copy:

- App name (the URL slug, not the display name)
- Client ID
- The downloaded private key PEM file

## Apps Acting on Behalf of Users

App settings:

- Generate a client secret
- Callback URL: `https://ADDRESS/callback`\
  (`ADDRESS` is Thrippy's public address for HTTP webhooks - see the `server` command's `-w` flag)

Details to copy:

- Client ID
- Client secret

## References

- [About creating GitHub Apps](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps)
- [Registering a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app)
- [Managing private keys for GitHub Apps](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/managing-private-keys-for-github-apps)
