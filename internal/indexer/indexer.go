package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

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

// fetchLatestDelegation fetches the most recent delegation from TzKT
func (i *Indexer) fetchLatestDelegation() (*models.Delegation, error) {
	response, err := http.Get(i.tzktURL + "?limit=1&sort.desc=id")
	if err != nil {
		return nil, fmt.Errorf("failed to get latest delegation: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var delegations []models.Delegation
	err = json.Unmarshal(body, &delegations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if len(delegations) == 0 {
		return nil, fmt.Errorf("no delegations found")
	}

	return &delegations[0], nil
}

// Poll fetches new delegations from TzKT and inserts them into the database
func (i *Indexer) Poll(ctx context.Context) error {
	log.Println("Polling for new delegations...")
	// TODO: Query TzKT API with id.gt=<cursor>&limit=100&sort.asc=id
	// TODO: Parse JSON response
	// TODO: Insert delegations using db.InsertDelegations
	// TODO: Update cursor to max id from batch
	return nil
}
