# Thrippy Link Setup: `jira-user-token`

## Atlassian User Account API Tokens

<https://id.atlassian.com/manage-profile/security/api-tokens>

> [!IMPORTANT]
> If you create the API token with specific [scopes](https://developer.atlassian.com/cloud/jira/platform/scopes-for-oauth-2-3LO-and-forge-apps/), specify **at least** `read:me` (classic), `read:jira-user` (classic), or `read:user:jira` (granular).

## Thrippy Link Setup

1. Create the link

   ```shell
   thrippy create-link --template jira-user-token
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> \
           --kv "base_url=https://your-domain.atlassian.net" \
           --kv "email=you@example.com" --kv "api_token=..."
   ```

## References

- [Manage API tokens for your Atlassian account](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/)
