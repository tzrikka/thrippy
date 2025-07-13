# Thrippy Link Setup: `bitbucket-user-token`

## Atlassian User Account API Tokens

https://id.atlassian.com/manage-profile/security/api-tokens

> [!IMPORTANT]
> If you create the API token with specific [scopes](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#forge-app-and-api-token-scopes), specify **at least** `read:user:bitbucket`.

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template bitbucket-user-token
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "email=..." --kv "api_token=..."
   ```

## References

- [Manage API tokens for your Atlassian account](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/)
- [BitBucket API tokens](https://support.atlassian.com/bitbucket-cloud/docs/api-tokens/)
- [BitBucket REST API authentication methods](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#api-tokens)

