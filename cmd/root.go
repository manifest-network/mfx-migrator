package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/utils"
)

// Design notes:
// - The `claim` command should be used to claim a work item. If no work items are available, the command should exit
// - The `verify` command should be used to verify the status of a migration of MFX tokens to the Manifest Ledger
// - The `migrate` command should be used to execute a migration of MFX tokens to the Manifest Ledger
// - The `*.uuid` files should be used to persist the state of a migration locally. This file should be removed once the migration is complete
//
// 0. Check if any `*.uuid` files exist in the current directory. If so, resume the migration
// 1. Claim a work item. Exit if no work items are available
//   1.1. If the work item is claimed successfully, the `*.uuid` file should be created.
//  	  Migration state should be persisted to this file.
//   1.2. If the work item is not claimed successfully, try the next work item
// 2. Verify the work item is valid
// 3. Execute the migration
// 4. Verify the migration was successful
// 5. POST the 'talib/complete-work/' endpoint to complete the work item
//   5.1. If the work item is completed, the `*.uuid` file should be removed
//        Note: Completed involves both successful and failed migrations.
//              Failed migrations should have a reason for failure persisted to the database.

var rootCmd = &cobra.Command{
	Use:   "mfx-migrator",
	Short: "Migrate your MFX tokens to the Manifest Ledger",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logLevelArg := viper.GetString("logLevel")
		urlString := viper.GetString("url")
		if err := setLogLevel(logLevelArg); err != nil {
			return err
		}
		if err := validateURL(urlString); err != nil {
			return err
		}

		slog.Debug("Application initialized", "logLevel", logLevelArg, "url", urlString)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("No config file found")
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	validLogLevels = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	validLogLevelsStr = strings.Join(utils.GetKeys(validLogLevels), "|")
)

func init() {
	rootCmd.PersistentFlags().StringP("logLevel", "l", "info", fmt.Sprintf("set log level (%s)", validLogLevelsStr))
	err := viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("logLevel"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	rootCmd.PersistentFlags().StringP("url", "u", "", "Root URL of the API server")
	err = viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	viper.AddConfigPath("./")
	viper.SetConfigName("config")

	viper.AutomaticEnv()
}

// setLogLevel sets the log level
func setLogLevel(logLevel string) error {
	level, exists := validLogLevels[logLevel]
	if !exists {
		return fmt.Errorf("invalid log level: %s. Valid log levels are: %s", logLevel, validLogLevelsStr)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	return nil
}

// validateURL validates a URL is not empty and is a valid URL
func validateURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("URL cannot be empty")
	}

	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	return nil
}
