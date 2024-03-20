package config

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"

	"github.com/liftedinit/mfx-migrator/internal/utils"
)

// Config represents the configuration for the command
type Config struct {
	UUID         string
	Url          string
	Neighborhood uint64
}

// Print the Config, omits the password
func (c Config) Print() {
	fmt.Printf("UUID: %v\n", c.UUID)
	fmt.Printf("Url: %v\n", c.Url)
	fmt.Printf("Neighborhood: %v\n", c.Neighborhood)
}

// Validate the Config making sure all required fields are present and valid
func (c Config) Validate() error {
	if c.Url == "" {
		return fmt.Errorf("url is required")
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

type AuthConfig struct {
	Username string // The username to authenticate with
	Password string // The password to authenticate with
}

func (c AuthConfig) Print() {
	fmt.Printf("Username: %v\n", c.Username)
}

func (c AuthConfig) Validate() error {
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	if c.Password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}

type ClaimConfig struct {
	Force bool // Force re-claiming of a failed work item
	Jobs  uint // Number of parallel jobs to claim
}

func (c ClaimConfig) Validate() error {
	if c.Jobs == 0 {
		return fmt.Errorf("jobs > 0 is required")
	}

	return nil
}

type MigrateConfig struct {
	ChainID          string                     // The destination chain ID
	AddressPrefix    string                     // The destination address prefix
	NodeAddress      string                     // The destination RPC node address
	KeyringBackend   string                     // The destination chain keyring backend to use
	BankAddress      string                     // The destination chain address of the bank account to send tokens from
	ChainHome        string                     // The root directory of the destination chain configuration
	TokenMap         map[string]utils.TokenInfo // Map of source token address to destination token info
	WaitTxTimeout    uint                       // Number of seconds spent waiting for the transaction to be included in a block
	WaitBlockTimeout uint                       // Number of seconds spent waiting for the block to be committed
}

func (c MigrateConfig) Validate() error {
	if c.ChainID == "" {
		return fmt.Errorf("chain ID is required")
	}

	if c.AddressPrefix == "" {
		return fmt.Errorf("address prefix is required")
	}

	if c.NodeAddress == "" {
		return fmt.Errorf("node address is required")
	}

	if c.KeyringBackend == "" {
		return fmt.Errorf("keyring backend is required")
	}

	if c.BankAddress == "" {
		return fmt.Errorf("bank address is required")
	}

	if c.ChainHome == "" {
		return fmt.Errorf("chain home is required")
	}

	if c.WaitTxTimeout == 0 {
		return fmt.Errorf("wait for tx timeout > 0 is required")
	}

	if c.WaitBlockTimeout == 0 {
		return fmt.Errorf("wait for block timeout > 0 is required")
	}

	return nil
}
