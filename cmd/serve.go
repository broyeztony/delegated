package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  `Start the HTTP server to serve delegation data via REST API.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		log.Println("Serve command started")
		fmt.Println("TODO: Implement API server")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
