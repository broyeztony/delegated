package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/broyeztony/delegated/internal/db"
	"github.com/broyeztony/delegated/internal/indexer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var backfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Backfill historical delegation data",
	Long:  `Backfills historical delegations from TzKT API using COPY protocol.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Backfill command started")

		// Initialize database connection
		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
		if err != nil {
			return fmt.Errorf("unable to create connection pool: %w", err)
		}
		defer dbpool.Close()

		// Get min ID in our table - if empty, exit
		ctx := context.Background()
		count, minID, err := db.GetMinID(ctx, dbpool)
		if err != nil {
			return fmt.Errorf("failed to get min id: %w", err)
		}

		if count == 0 {
			log.Println("Delegations table is empty. Run 'index' command first to populate recent delegations, then run backfill.")
			return fmt.Errorf("cannot backfill: table `delegations` is empty")
		}

		// Start backfill
		log.Printf("Starting backfill from cursor: %d (oldest ID in table)\n", minID)

		idx := indexer.NewIndexer(dbpool)
		startTime := time.Now()

		totalRecords, totalBatches, err := idx.Backfill(ctx, minID)
		if err != nil {
			return err
		}

		// Print summary
		totalDuration := time.Since(startTime)
		log.Printf("\nBackfill Summary:")
		log.Printf("Total records: %d", totalRecords)
		log.Printf("Total batches: %d", totalBatches)
		log.Printf("Total duration: %v", totalDuration)
		if totalRecords > 0 {
			log.Printf("Avg records/sec: %.2f", float64(totalRecords)/totalDuration.Seconds())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(backfillCmd)
}
