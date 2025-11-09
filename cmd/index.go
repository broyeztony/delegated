package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/broyeztony/delegated/internal/indexer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var (
	pollingInterval int
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Start indexing delegations",
	Long:  `Continuously poll and index new Tezos delegations from tzkt.io API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Index command started")

		// Initialize database connection
		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
		if err != nil {
			return fmt.Errorf("unable to create connection pool: %w", err)
		}
		defer dbpool.Close()

		// Create indexer
		idx := indexer.NewIndexer(dbpool)

		// Initialize cursor
		ctx := context.Background()
		if err := idx.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		// Start polling loop (every 60s)
		ticker := time.NewTicker(time.Duration(pollingInterval) * time.Second)
		defer ticker.Stop()

		// Initial poll
		if err := idx.Poll(ctx); err != nil {
			log.Printf("Error in initial poll: %v\n", err)
		}

		for range ticker.C {
			if err := idx.Poll(ctx); err != nil {
				log.Printf("Error polling: %v\n", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.Flags().IntVarP(&pollingInterval, "interval", "i", 60, "Polling interval in seconds")
}
