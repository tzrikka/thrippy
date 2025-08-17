# Thrippy

[![Go Reference](https://pkg.go.dev/badge/github.com/tzrikka/thrippy.svg)](https://pkg.go.dev/github.com/tzrikka/thrippy)
[![Go Report Card](https://goreportcard.com/badge/github.com/tzrikka/thrippy)](https://goreportcard.com/report/github.com/tzrikka/thrippy)

Thrippy is a CLI application and gRPC client/server to manage authentication configurations and secret tokens for third-party ("3P") services.

It supports both static and OAuth 2.0 credentials, and it is designed to be simple and secure by default.

## Overview

Thrippy manages "links", which are collections of configurations, credentials, and metadata.

When you create a link, you specify a "template" for it, which identifies a specific well-known service (e.g. ChatGPT, GitHub, Gmail, Slack) and its authentication type (see the list below). This enables Thrippy to set most configuration details automatically.

Static credentials (e.g. API keys) are set manually by the user. Dynamic credentials (e.g. OAuth 2.0 tokens) are refreshed automatically by Thrippy after an initial interactive user authorization.

## Supported Services and Auth Types

- [Atlassian](./docs/atlassian/README.md)
  - Products: Bitbucket, Confluence, Jira
  - OAuth 2.0 (3LO) app / user API token / webhook-only
- [ChatGPT](./docs/chatgpt/README.md)
  - Static API key
- [Claude](./docs/claude/README.md)
  - Static API key
- [Generic OAuth 2.0](./docs/generic-oauth/README.md)
- [GitHub](./docs/github/README.md)
  - App installation using JWTs based on static credentials
  - App authorization to act on behalf of a user
  - User's static Personal Access Token (PAT)
  - Webhook
- [Google](./docs/google/README.md)
  - OAuth 2.0 to act on behalf of a user
  - Static Google Cloud service account key
  - Gemini using a static API key
- [Slack](./docs/slack/README.md)
  - App using a static bot token
  - App using OAuth v2 (regular Slack / GovSlack)
  - Private "Socket Mode" app using a static app-level token

## Quickstart

1. Install Thrippy:

   ```shell
   go install github.com/tzrikka/thrippy
   ```

> [!TIP]
> The binary will be located here: `$(go env GOPATH)/bin`

2. Start the Thrippy server:

   ```shell
   thrippy server --dev
   ```

> [!IMPORTANT]
> In dev mode, Thrippy uses an in-memory secrets manager by default, which is destroyed when the server goes down.

3. Create any **static** link, based on the [documentation](https://github.com/tzrikka/thrippy/tree/main/docs)

## Production Server Configuration

- Secure secrets manager
- [HTTP tunnel to enable OAuth 2.0 links](./docs/http_tunnel.md)
- [m/TLS for Thrippy client/server communication](./x509/README.md)
