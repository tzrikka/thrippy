# HTTP Tunneling

Thrippy includes a local HTTP server to support OAuth 2.0, and it requires a public HTTPS URL to receive OAuth 2.0 redirects.

This is a simple and free way to set it up.

## _ngrok_

1. Sign up for an _ngrok_ account: <https://dashboard.ngrok.com/signup>

2. Copy your _ngrok_ auth token from: <https://dashboard.ngrok.com/get-started/your-authtoken>

3. Install and set up the _ngrok_ CLI agent: <https://dashboard.ngrok.com/get-started/setup>

4. Create a free static domain in: <https://dashboard.ngrok.com/domains>

5. Run this command to start an HTTP tunnel:

   ```shell
   ngrok http --url=xxx-yyy-zzz.ngrok-free.app 14470
   ```

   - `xxx-yyy-zzz.ngrok-free.app` is the free static domain from step 4
   - `14470` is the local port of the Thrippy HTTP server

## Thrippy Configuration

When starting Thrippy, add this flag:

```shell
thrippy server --webhook-addr xxx-yyy-zzz.ngrok-free.app
```

Or set it with an environment variable:

```shell
export THRIPPY_WEBHOOK_ADDRESS=xxx-yyy-zzz.ngrok-free.app
```

Or add it to the file `$XDG_CONFIG_HOME/thrippy/config.toml`:

```toml
[server]
webhook_address = "xxx-yyy-zzz.ngrok-free.app"
```

(If `$XDG_CONFIG_HOME` isn't set, the default path per OS is specified [here](https://github.com/tzrikka/xdg/blob/main/README.md#default-paths)).
