# Thrippy Link Setup: `bitbucket-app-oauth`

## Bitbucket App Minimum Settings

## Bitbucket App Details to Copy

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template bitbucket-app-oauth \
           --client-id "..." --client-secret "..." \
           [ --scopes "xxx,yyy,..." [ --scopes "zzz" ] ]
   ```

2. Install and authorize the Bitbucket app (interactively in a browser)

   ```shell
   thrippy start-oauth <link ID>
   ```

## References

- [Bitbucket Cloud OAuth 2.0](https://developer.atlassian.com/cloud/bitbucket/oauth-2/)
- [Use OAuth on Bitbucket Cloud](https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/)
