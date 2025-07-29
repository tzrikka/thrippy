# Thrippy Link Setup: `bitbucket-webhook`

All Bitbucket link templates support webhooks for inbound events, as an optional addition.

This link template defines **only** a webhook, without outbound API access.

This can be for any type of event (repository, issue, pull request).

1. Create the link

   ```shell
   thrippy create-link --template bitbucket-webhook
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "webhook_secret=..."
   ```

## References

- [Manage webhooks](https://support.atlassian.com/bitbucket-cloud/docs/manage-webhooks/)
- [Event payloads](https://support.atlassian.com/bitbucket-cloud/docs/event-payloads/)
