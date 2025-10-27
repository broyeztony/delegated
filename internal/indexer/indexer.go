package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/broyeztony/delegated/internal/db"
	"github.com/broyeztony/delegated/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Indexer struct {
	pool    *pgxpool.Pool
	cursor  int64
	tzktURL string
}

func NewIndexer(pool *pgxpool.Pool) *Indexer {
	return &Indexer{
		pool:    pool,
		tzktURL: "https://api.tzkt.io/v1/operations/delegations",
	}
}

// Initialize sets up the cursor (latest TzKT delegation if table empty)
func (i *Indexer) Initialize(ctx context.Context) error {
	count, maxID, err := db.GetMaxID(ctx, i.pool)
	if err != nil {
		return fmt.Errorf("failed to get max id: %w", err)
	}

	if count == 0 {
		log.Println("Table is empty, fetching latest delegation from TzKT...")
		latestDelegation, err := i.fetchLatestDelegation()
		if err != nil {
			return fmt.Errorf("failed to fetch latest delegation: %w", err)
		}

		// Insert the latest delegation into the database
		if err := db.BulkInsertDelegations(ctx, i.pool, []models.Delegation{*latestDelegation}); err != nil {
			return fmt.Errorf("failed to insert latest delegation: %w", err)
		}
		log.Println("Inserted latest delegation into database")

		// Update cursor only after successful insert
		i.cursor = latestDelegation.ID
		log.Printf("Latest delegation id: %d\n", i.cursor)
	} else {
		log.Printf("Resuming from id: %d\n", maxID)
		i.cursor = maxID
	}

	return nil
}

// fetchDelegations fetches delegations from TzKT with given query parameters
func (i *Indexer) fetchDelegations(queryParams string) ([]models.Delegation, error) {
	response, err := http.Get(i.tzktURL + queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch delegations: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var delegations []models.Delegation
	err = json.Unmarshal(body, &delegations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return delegations, nil
}

// fetchLatestDelegation fetches the most recent delegation from TzKT
func (i *Indexer) fetchLatestDelegation() (*models.Delegation, error) {
	delegations, err := i.fetchDelegations("?limit=1&sort.desc=id")
	if err != nil {
		return nil, err
	}

	if len(delegations) == 0 {
		return nil, fmt.Errorf("no delegations found")
	}

	return &delegations[0], nil
}

func (i *Indexer) fetchNewDelegations(ctx context.Context, cursor int64) ([]models.Delegation, error) {
	query := "?id.gt=" + strconv.FormatInt(cursor, 10) + "&limit=100&sort.asc=id"
	return i.fetchDelegations(query)
}

// FetchNewDelegations is a public method for fetching delegations with a specific cursor (used by backfill)
func (i *Indexer) FetchNewDelegations(ctx context.Context, cursor int64) ([]models.Delegation, error) {
	// Use id.lt to go backward in time (older records have smaller IDs)
	// Sort desc to get the most recent records within the range
	query := "?id.lt=" + strconv.FormatInt(cursor, 10) + "&limit=10000&sort.desc=id"
	return i.fetchDelegations(query)
}

// Backfill fetches historical delegations going backward from the given cursor and inserts directly into delegations using COPY protocol
func (i *Indexer) Backfill(ctx context.Context, startCursor int64) (totalRecords int, totalBatches int, err error) {
	cursor := startCursor

	for {
		log.Printf("Fetching batch %d (cursor: %d)\n", totalBatches+1, cursor)

		delegations, err := i.FetchNewDelegations(ctx, cursor)
		if err != nil {
			return totalRecords, totalBatches, fmt.Errorf("failed to fetch: %w", err)
		}

		if len(delegations) == 0 {
			log.Println("No more delegations found. Backfill complete!")
			break
		}

		insertStart := time.Now()
		if err := db.CopyInsertDelegations(ctx, i.pool, delegations); err != nil {
			return totalRecords, totalBatches, fmt.Errorf("failed to copy: %w", err)
		}
		insertDuration := time.Since(insertStart)

		totalBatches++
		totalRecords += len(delegations)
		log.Printf("Inserted %d records in %v (total: %d records in %d batches)\n",
			len(delegations), insertDuration, totalRecords, totalBatches)

		cursor = delegations[len(delegations)-1].ID

		time.Sleep(100 * time.Millisecond)
	}

	return totalRecords, totalBatches, nil
}

// Poll fetches new delegations from TzKT and inserts them into the database
func (i *Indexer) Poll(ctx context.Context) error {
	log.Println("Polling for new delegations...")

	newDelegations, err := i.fetchNewDelegations(ctx, i.cursor)
	if err != nil {
		return fmt.Errorf("failed to fetch new delegations: %w", err)
	}

	if len(newDelegations) == 0 {
		log.Println("No new delegations found")
		return nil
	}

	log.Println("Found", len(newDelegations), "new delegations")

	if err := db.BulkInsertDelegations(ctx, i.pool, newDelegations); err != nil {
		return fmt.Errorf("failed to insert new delegations: %w", err)
	}

	log.Println("Inserted", len(newDelegations), "new delegations into database")

	// Update cursor only after successful insert
	i.cursor = newDelegations[len(newDelegations)-1].ID
	log.Printf("Updated cursor to: %d\n", i.cursor)

	return nil
}
