package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	dbURL   string
)

var rootCmd = &cobra.Command{
	Use:   "delegated",
	Short: "Tezos delegation indexer and API server",
	Long:  `A simple service to index and serve Tezos delegation data.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.solution1.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbURL, "db-url", "", "database connection URL (default from DB_URL env var)")

	viper.BindPFlag("db-url", rootCmd.PersistentFlags().Lookup("db-url"))
	viper.SetEnvPrefix("")
	viper.BindEnv("db-url", "DB_URL")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".delegated")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// getDatabaseURL retrieves the database connection string from flag, env var, or returns error if not set
func getDatabaseURL() (string, error) {
	connStr := viper.GetString("db-url")

	if connStr == "" {
		connStr = os.Getenv("DB_URL")
	}

	if connStr == "" {
		return "", fmt.Errorf("database URL is required: set --db-url flag or DB_URL environment variable")
	}

	return connStr, nil
}
