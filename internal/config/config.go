package config

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/google/uuid"

	"github.com/manifest-network/mfx-migrator/internal/utils"
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
	Binary           string                     // Binary name of the destination blockchain
	GasPrice         float64                    // Minimum gas price to use for transactions
	GasAdjustment    float64                    // Gas adjustment to use for transactions
	GasDenom         string                     // Gas denomination to use for transactions
	FeeGranter       string                     // The address of the gas fee granter
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

	if c.Binary == "" {
		return fmt.Errorf("binary is required")
	}

	if c.GasPrice < 0 {
		return fmt.Errorf("gas price must be >= 0")
	}

	if c.GasAdjustment < 0 {
		return fmt.Errorf("gas adjustment must be >= 0")
	}

	if c.GasDenom == "" {
		return fmt.Errorf("gas denom is required")
	}

	if c.FeeGranter == "" {
		return fmt.Errorf("fee granter is required")
	}

	if _, err := exec.LookPath(c.Binary); err != nil {
		return fmt.Errorf("binary %s not found in PATH", c.Binary)
	}

	return nil
}
