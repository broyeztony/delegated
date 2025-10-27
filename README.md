# Delegated: Simple Tezos Delegation Indexer And Data API

## Requirements

- Go 1.23 or above
- PostgreSQL

## Build

```bash
# Option 1: Build to bin folder
go build -o bin/delegated

# Option 2: Install to GOPATH/bin
go install
```

## Setup

```bash
# Set database connection URL
export DB_URL="postgresql://localhost/delegated"

# Create database
createdb delegated

# Load schema
psql delegated < schema.sql
```

## Verify

```bash
# Connect to database
psql delegated

# Check table structure
\d delegations


                        Table "public.delegations"
  Column   |            Type             | Collation | Nullable | Default 
-----------+-----------------------------+-----------+----------+---------
 id        | bigint                      |           | not null | 
 delegator | character varying(36)       |           | not null | 
 timestamp | timestamp without time zone |           | not null | 
 amount    | bigint                      |           | not null | 
 level     | integer                     |           | not null | 
Indexes:
    "delegations_pkey" PRIMARY KEY, btree (id)
    "idx_delegations_timestamp" btree ("timestamp")


# Query table
SELECT * FROM delegations ORDER BY timestamp DESC LIMIT 10;
\q
```

## Usage

### Start Indexer

```bash
# Terminal 1: Start indexing delegations (runs every 60s)
export DB_URL="postgresql://localhost/delegated"
./bin/delegated index
# OR if installed via go install:
delegated index
```

### Backfill Historical Data

```bash
# Terminal 2:Run backfill to populate historical delegations (requires delegations table to have at least 1 record as a starting point for backfilling)
export DB_URL="postgresql://localhost/delegated"
./bin/delegated backfill
# OR if installed via go install:
delegated backfill
```

**Note:** The backfill command requires the delegations table to have at least one record (run `index` first). It fetches historical delegations going back to the earliest delegation in June 2018 and stores them in the `delegations` table using COPY protocol for performance.

### Start API Server

```bash
# Terminal 3: Start the API server
export DB_URL="postgresql://localhost/delegated"
./bin/delegated serve
# OR if installed via go install:
delegated serve
```

### Query API

```bash
# Get all delegations
curl http://localhost:8080/xtz/delegations | jq

# Filter by year
curl http://localhost:8080/xtz/delegations?year=2022 | jq
```

**Example Response:**
```json
{
  "data": [
    {
      "timestamp": "2025-10-26T17:17:52Z",
      "amount": "161512757",
      "delegator": "tz1cAuZvhNgybyXdu4x2263CNjaFqHKdd8eo",
      "level": "10674288"
    },
    {
      "timestamp": "2025-10-26T17:03:08Z",
      "amount": "1287400959",
      "delegator": "tz1QMwsi5onCV9yR2VJsSCRBDsrSfzrViMss",
      "level": "10674179"
    }
  ]
}
```

## My approach

### Tezos API exploration

I reviewed the API spec at https://api.tzkt.io/#operation/Operations_GetDelegations and ran a couple of queries to determine:
 - how often new delegations are registered, to estimate the polling interval for live ingestion and the size of the window we need to fetch. The interval is set to 60 seconds. From what I experienced, many intervals will be empty (no new delegations)
 - what was the first delegation registered, to estimate how far we will need to go for backfilling historical data and how many records that will represent. The first delegation returned by the API has the timestamp 2018-06-30T19:30:27Z and there are a little more than 771000 delegations registered to date
 - how many records we can fetch at once. It turns out that 10,000 is the maximum number of records we can fetch in a single query. Trying to query more than 10,000 will return 
 ```
 {
  "code": 400,
  "errors": {
    "limit": "The field limit must be between 0 and 10000."
  }
}
```

 ### Live indexing

 Delegations are associated with a monotonically increasing `id` field, which belongs to the sortable fields set. 
 We use this field to fetch newer (and older) delegations.
 
 The number of new delegations returned for each poll is relatively small. 
 We insert them in database (Postgresql) using a rollable bulk insert transaction.
 
 If that transaction succeed, we update the in-memory `cursor` value which is the highest `id` we just inserted. 
 
 The live indexing is resilient. If an error occurs, it can retry. If the app is terminated, it can be resumed later and catch up.

### Backfilling

We fetch historical delegations by the maximum batch size permitted by the API (limit=10000) using the `id` field to work backward from the oldest record already present in our local database down to the oldest delegations returned by the Tezos API. 

We continue until the API returns zero records.

#### Performance Optimization: Direct COPY Protocol

We use PostgreSQL's [COPY protocol](https://www.postgresql.org/docs/current/sql-copy.html) for bulk insertion, achieving ~18,800 records/second. On a MacBook Pro M1 2021 with 16GB Memory, 771,000+ records are backfilled in ~40 seconds across 78 batches.

```bash
# Verify data integrity after backfill
psql delegated -c "SELECT 
  COUNT(*) as total_records,
  COUNT(CASE WHEN delegator = '' THEN 1 END) as empty_delegator,
  COUNT(CASE WHEN timestamp IS NULL THEN 1 END) as null_timestamp,
  COUNT(CASE WHEN amount IS NULL THEN 1 END) as null_amount,
  COUNT(CASE WHEN level IS NULL THEN 1 END) as null_level
FROM delegations;"
```

Result: 0 empty strings, 0 NULL values across 771,332 records.

#### Technical Decision: Direct COPY vs. Staging Table

We chose to COPY directly into the target `delegations` table (which has a PRIMARY KEY constraint on `id` and an index on `timestamp`) rather than using a staging table approach. This decision is justified by:

1. **Data Quality**: After testing 771,332 records, we found zero duplicates or inconsistencies in the TzKT API data
2. **Performance**: Constraints don't significantly impact COPY performance - we still achieve 18k+ records/sec
3. **Simplicity**: Eliminates staging-to-target transfer step and reduces complexity

**Trade-off**: If a duplicate somehow appears, the entire 10k records batch fails. However, since we verified no duplicates exist in the source data over 771k+ records, the risk is acceptable for the significant performance gains, in the context of a tech assignment, executed on the local machine.

**Note**: PostgreSQL supports `ON_ERROR ignore` for error handling in COPY, but pgx (our PostgreSQL driver) doesn't support it yet ([issue #1362](https://github.com/jackc/pgx/issues/1362)). Future enhancement could add per-row error handling.


