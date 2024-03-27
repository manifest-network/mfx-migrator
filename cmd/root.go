package cmd

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/utils"
)

var rootCmd = &cobra.Command{
	Use:               "mfx-migrator",
	Short:             "Migrate your MFX tokens to the Manifest Ledger",
	PersistentPreRunE: RootCmdPersistentPreRunE,
}

func RootCmdPersistentPreRunE(cmd *cobra.Command, args []string) error {
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
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("Using config file", "file", viper.ConfigFileUsed())
	} else {
		slog.Info("No config file found")
	}

	if err := rootCmd.Execute(); err != nil {
		slog.Error("An error occurred", "error", err)
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

func SetupRootCmdFlags(command *cobra.Command) {
	command.PersistentFlags().StringP("logLevel", "l", "info", fmt.Sprintf("set log level (%s)", validLogLevelsStr))
	if err := viper.BindPFlag("logLevel", command.PersistentFlags().Lookup("logLevel")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
	}

	command.PersistentFlags().String("url", "", "Root URL of the API server")
	if err := viper.BindPFlag("url", command.PersistentFlags().Lookup("url")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
	}

	command.PersistentFlags().Uint64("neighborhood", 0, "Neighborhood ID")
	if err := viper.BindPFlag("neighborhood", command.PersistentFlags().Lookup("neighborhood")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
	}

	command.PersistentFlags().String("username", "", "Username for the remote database")
	if err := viper.BindPFlag("username", command.PersistentFlags().Lookup("username")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
	}

	command.PersistentFlags().String("password", "", "Password for the remote database")
	if err := viper.BindPFlag("password", command.PersistentFlags().Lookup("password")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
	}

	command.SilenceUsage = true
	command.SilenceErrors = true
}

func init() {
	SetupRootCmdFlags(rootCmd)

	viper.AddConfigPath("./")
	viper.AddConfigPath("/config")
	viper.SetConfigName("migrator-config")

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
		return fmt.Errorf("URL cannot be empty")
	}

	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	return nil
}
