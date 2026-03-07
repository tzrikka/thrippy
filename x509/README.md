# X.509 Certificates for gRPC

This directory enables TLS and mTLS in gRPC connections, based on: \
<https://github.com/grpc/grpc-go/blob/master/examples/data/x509/README.md>

## Option 1: Unauthenticated & Unencrypted

This is supported only for the sake of simplicity during development and tests. It is unsafe in production!

This is enabled only when both server and client use the `--dev` flag.

## Option 2: TLS

In this mode, all gRPC traffic is encrypted, but only the server identity is authenticated.

1. Create test certificates

   ```shell
   ./create_server.sh
   ```

2. Configure the Thrippy server, using its public and private keys (in the file `${XDG_CONFIG_HOME}/thrippy/config.toml`):

   ```toml
   [grpc.server]
   server_cert = "<absolute path>/thrippy_server_cert.pem"
   server_key = "<absolute path>/thrippy_server_key.pem"
   ```

3. Configure the Thrippy client, using the server's CA public key (also in the file `${XDG_CONFIG_HOME}/thrippy/config.toml`):

   ```toml
   [grpc.client]
   server_ca_cert = "<absolute path>/thrippy_server_ca_cert.pem"
   server_name_override = "thrippy.example.com"
   ```

> [!NOTE]
> Clients may or may not be on the same computer as the server, i.e. you may configure both the `[grpc.server]` and the `[grpc.client]` sections in the same `config.toml` file.

## Option 3: Mutual TLS (mTLS)

In this mode, all gRPC traffic is encrypted, and the identities of both sides are authenticated.

1. Create test certificates

   ```shell
   ./create_server.sh && ./create_client.sh
   ```

2. Configure the Thrippy server, using its public and private keys, and the **client's** CA public key (in the file `${XDG_CONFIG_HOME}/thrippy/config.toml`):

   ```toml
   [grpc.server]
   client_ca_cert = "<absolute path>/thrippy_client_ca_cert.pem"
   server_cert = "<absolute path>/thrippy_server_cert.pem"
   server_key = "<absolute path>/thrippy_server_key.pem"
   ```

3. Configure the Thrippy client, using its public and private keys, and the **server's** CA public key (also in the file `${XDG_CONFIG_HOME}/thrippy/config.toml`):

   ```toml
   [grpc.client]
   client_cert = "<absolute path>/thrippy_client_cert.pem"
   client_key = "<absolute path>/thrippy_client_key.pem"
   server_ca_cert = "<absolute path>/thrippy_server_ca_cert.pem"
   server_name_override = "thrippy.example.com"
   ```

> [!NOTE]
> Clients may or may not be on the same computer as the server, i.e. you may configure both the `[grpc.server]` and the `[grpc.client]` sections in the same `config.toml` file.
