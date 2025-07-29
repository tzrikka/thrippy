# Thrippy Link Setup: `github-webhook`

All GitHub link templates support webhooks for inbound events, as an optional addition.

This link template defines **only** a webhook, without outbound API access.

This can be for any type of GitHub event (app, organization, repository, GitHub Marketplace, or GitHub Sponsors), using any type of payload (JSON or web form).

1. Create the link

   ```shell
   thrippy create-link --template github-webhook
   ```

2. Set the link's static credentials

   ```shell
   thrippy set-creds <link ID> --kv "webhook_secret=..."
   ```

## References

- [Webhooks documentation](https://docs.github.com/webhooks)
- [Creating webhooks](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks)
- [Handling webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/handling-webhook-deliveries)
- [Building a GitHub App that responds to webhook events](https://docs.github.com/en/apps/creating-github-apps/writing-code-for-a-github-app/building-a-github-app-that-responds-to-webhook-events)
