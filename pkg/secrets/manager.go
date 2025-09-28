// Package secrets provides a generic interface for managing
// user secrets, using one of these providers:
//   - In-memory storage ("in-memory") - see note below!
//   - AWS Secrets Manager ("aws")
//   - Google Cloud Secret Manager ("gcp")
//   - HashiCorp Vault ("vault")
//   - Infisical ("infisical")
//
// Configuration in environment variables:
//   - THRIPPY_SECRETS_PROVIDER
//   - THRIPPY_SECRETS_NAMESPACE
//   - VAULT_ADDR
//   - VAULT_CACERT
//   - VAULT_TOKEN
//
// Configuration in the file "$XDG_CONFIG_HOME/thrippy/config.toml":
//
//	[secrets]
//	provider = "in-memory"
//	namespace = "default"
//
//	[secrets.vault]
//	address = "https://127.0.0.1:8200"
//	cacert = "/path/to/vault-ca.pem"
//	token = "..."
//
// Notes:
//   - The in-memory provider is used by default when specifying the "--dev" flag,
//     but it is unreliable and insecure for real-world use!
package secrets

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	defaultProvider  = inMemoryOption
	defaultNamespace = "default" // Other examples: "dev", "staging", "prod", etc.
)

// ManagerFlags defines global (but hidden) CLI flags. The purpose
// of these CLI flags is to initialize the generic secrets manager via
// environment variables and/or the application's configuration file.
func ManagerFlags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "secrets-provider",
			Value: defaultProvider,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_SECRETS_PROVIDER"),
				toml.TOML("secrets.provider", configFilePath),
			),
			Hidden: true,
			Validator: func(v string) error {
				options := map[string]bool{
					fileOption:     true,
					inMemoryOption: true,
					vaultOption:    true,
				}
				if ok := options[v]; !ok {
					return errors.New("unrecognized option")
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:  "secrets-namespace",
			Value: defaultNamespace,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_SECRETS_NAMESPACE"),
				toml.TOML("secrets.namespace", configFilePath),
			),
			Hidden: true,
		},
	}
}

// Manager is a simple interface to manage user secrets.
type Manager interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type genericWrapper struct {
	provider  Manager
	namespace string
}

func NewManager(cmd *cli.Command) (Manager, error) {
	provider := cmd.String("secrets-provider")
	ns := cmd.String("secrets-namespace")

	if provider == defaultProvider && !cmd.Bool("dev") {
		return nil, errors.New("in-memory secrets provider allowed only with --dev flag")
	}

	log.Info().Msgf("secrets provider: %s", provider)
	var p Manager
	var err error

	switch provider {
	case fileOption:
		p, err = newFileProvider()
	case inMemoryOption:
		p, err = newInMemoryProvider()
	case vaultOption:
		p, err = newVaultProvider(cmd)
	default:
		return nil, fmt.Errorf("unrecognized secrets provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}
	return &genericWrapper{provider: p, namespace: ns}, nil
}

// NewTestManager should be used only in unit tests.
func NewTestManager() Manager {
	p, _ := newInMemoryProvider()
	return &genericWrapper{provider: p, namespace: "test"}
}

func (m *genericWrapper) Set(ctx context.Context, key, value string) error {
	return m.provider.Set(ctx, m.namespaced(key), value)
}

func (m *genericWrapper) Get(ctx context.Context, key string) (string, error) {
	return m.provider.Get(ctx, m.namespaced(key))
}

func (m *genericWrapper) Delete(ctx context.Context, key string) error {
	return m.provider.Delete(ctx, m.namespaced(key))
}

func (m *genericWrapper) namespaced(key string) string {
	return fmt.Sprintf("thrippy/%s/%s", m.namespace, key)
}
