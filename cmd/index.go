package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Start indexing delegations",
	Long:  `Continuously poll and index new Tezos delegations from tzkt.io API.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		log.Println("Index command started")
		fmt.Println("TODO: Implement indexer logic")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
