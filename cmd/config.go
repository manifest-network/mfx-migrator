package cmd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/utils"
)

// Config represents the configuration for the command
type Config struct {
	Force        bool
	UUID         string
	Url          string
	Username     string
	Password     string
	Neighborhood uint64
}

// Print the Config, omits the password
func (c Config) Print() {
	fmt.Printf("Force: %v\n", c.Force)
	fmt.Printf("UUID: %v\n", c.UUID)
	fmt.Printf("Url: %v\n", c.Url)
	fmt.Printf("Username: %v\n", c.Username)
	fmt.Printf("Neighborhood: %v\n", c.Neighborhood)
}

// Validate the Config making sure all required fields are present and valid
func (c Config) Validate() error {
	if c.Username == "" {
		return errors.New("username is required")
	}

	if c.Password == "" {
		return errors.New("password is required")
	}

	if c.Url == "" {
		return errors.New("url is required")
	}

	if c.UUID != "" {
		_, err := uuid.Parse(c.UUID)
		if err != nil {
			return fmt.Errorf("could not parse UUID: %w", err)
		}
	}

	_, err := url.Parse(c.Url)
	if err != nil {
		return fmt.Errorf("could not parse URL: %w", err)
	}

	return nil
}

// LoadConfigFromCLI loads the Config from the CLI flags
//
// `uuidKey` is the name of the viper key that holds the UUID
// This is necessary because the UUID is used in multiple commands
func LoadConfigFromCLI(uuidKey string) Config {
	return Config{
		Force:        viper.GetBool("force"),
		UUID:         viper.GetString(uuidKey),
		Url:          viper.GetString("url"),
		Username:     viper.GetString("username"),
		Password:     viper.GetString("password"),
		Neighborhood: viper.GetUint64("neighborhood"),
	}
}

type MigrateConfig struct {
	ChainID        string                     // The destination chain ID
	AddressPrefix  string                     // The destination address prefix
	NodeAddress    string                     // The destination RPC node address
	KeyringBackend string                     // The destination chain keyring backend to use
	BankAddress    string                     // The destination chain address of the bank account to send tokens from
	ChainHome      string                     // The root directory of the destination chain configuration
	TokenMap       map[string]utils.TokenInfo // Map of source token address to destination token info
}

func LoadMigrationConfigFromCLI() MigrateConfig {
	var tokenMap map[string]utils.TokenInfo
	if err := viper.UnmarshalKey("token-map", &tokenMap); err != nil {
		panic(err)
	}
	return MigrateConfig{
		ChainID:        viper.GetString("chain-id"),
		AddressPrefix:  viper.GetString("address-prefix"),
		NodeAddress:    viper.GetString("node-address"),
		KeyringBackend: viper.GetString("keyring-backend"),
		BankAddress:    viper.GetString("bank-address"),
		ChainHome:      viper.GetString("chain-home"),
		TokenMap:       tokenMap,
	}
}

func (c MigrateConfig) Validate() error {
	if c.ChainID == "" {
		return errors.New("chain ID is required")
	}

	if c.AddressPrefix == "" {
		return errors.New("address prefix is required")
	}

	if c.NodeAddress == "" {
		return errors.New("node address is required")
	}

	if c.KeyringBackend == "" {
		return errors.New("keyring backend is required")
	}

	if c.BankAddress == "" {
		return errors.New("bank address is required")
	}

	if c.ChainHome == "" {
		return errors.New("chain home is required")
	}

	return nil
}
