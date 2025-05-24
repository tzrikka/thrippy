# Thrippy

[![Go Reference](https://pkg.go.dev/badge/github.com/tzrikka/thrippy.svg)](https://pkg.go.dev/github.com/tzrikka/thrippy)
[![Go Report Card](https://goreportcard.com/badge/github.com/tzrikka/thrippy)](https://goreportcard.com/report/github.com/tzrikka/thrippy)

Thrippy is a CLI application and gRPC client/server to manage authentication configurations and secret tokens for third-party (3P) services.

It supports both static and OAuth 2.0 credentials, and it is designed to be simple and secure by default.

## Supported Services and Auth Types

- [ChatGPT](./docs/chatgpt/README.md)
  - Static API key
- [Claude](./docs/claude/README.md)
  - Static API key
- [Generic OAuth 2.0](./docs/generic-oauth/README.md)
- [GitHub](./docs/github/README.md)
  - App installation using JWTs based on static credentials
  - App authorization to act on behalf of a user
  - User's static Personal Access Token (PAT)
- [Google](./docs/google/README.md)
  - OAuth 2.0 to act on behalf of a user
  - Static Google Cloud service account key
  - Gemini with a static API key
- [Slack](./docs/slack/README.md)
  - App using a static bot token
  - App using OAuth v2 (regular Slack / GovSlack)
