# Thrippy Link Setup: `confluence-user-token`

## Atlassian User Account API Tokens

https://id.atlassian.com/manage-profile/security/api-tokens

> [!IMPORTANT]
> If you create the API token with specific [scopes](https://developer.atlassian.com/cloud/confluence/scopes-for-oauth-2-3LO-and-forge-apps/), specify **at least** `read:me` (classic), `read:confluence-user` (classic), or `read:content-details:confluence` (granular).

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template confluence-user-token
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> \
           --kv "base_url=https://your-domain.atlassian.net" \
           --kv "email=..." --kv "api_token=..."
   ```

## References

- [Manage API tokens for your Atlassian account](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/)
