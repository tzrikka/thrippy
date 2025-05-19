package secrets

import (
	"context"
	"errors"

	vault "github.com/hashicorp/vault/api"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	vaultOption = "vault"
)

// VaultFlags defines global (but hidden) CLI flags. The purpose of
// these CLI flags is to initialize the HashiCorp Vault provider via
// environment variables and/or the application's configuration file.
func VaultFlags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "secrets-vault-address",
			Value: vault.DefaultAddress,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar(vault.EnvVaultAddress),
				toml.TOML("secrets.vault.address", configFilePath),
			),
			Hidden: true,
		},
		&cli.StringFlag{
			Name: "secrets-vault-cacert",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar(vault.EnvVaultCACert),
				toml.TOML("secrets.vault.cacert", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
		&cli.StringFlag{
			Name: "secrets-vault-token",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar(vault.EnvVaultToken),
				toml.TOML("secrets.vault.token", configFilePath),
			),
			Hidden: true,
		},
	}
}

type vaultProvider struct {
	client *vault.KVv2
}

func newVaultProvider(cmd *cli.Command) (Manager, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = cmd.String("secrets-vault-address")

	cacert := cmd.String("secrets-vault-cacert")
	if cacert != "" {
		if err := cfg.ConfigureTLS(&vault.TLSConfig{CACert: cacert}); err != nil {
			return nil, err
		}
	}

	client, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	client.SetToken(cmd.String("secrets-vault-token"))

	return &vaultProvider{client: client.KVv2("secret")}, nil
}

// The data size limit is 0.5 or 1 MiB, according to this link:
// https://developer.hashicorp.com/vault/docs/internals/limits.
func (p *vaultProvider) Set(ctx context.Context, key, value string) error {
	_, err := p.client.Put(ctx, key, map[string]any{"value": value})
	return err
}

func (p *vaultProvider) Get(ctx context.Context, key string) (string, error) {
	sec, err := p.client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	data, ok := sec.Data["value"].(string)
	if !ok {
		return "", errors.New("invalid data")
	}
	return data, nil
}

func (p *vaultProvider) Delete(ctx context.Context, key string) error {
	return p.client.DeleteMetadata(ctx, key)
}
