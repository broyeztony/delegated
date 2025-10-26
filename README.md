# Solution 1: Simple Tezos Delegation Indexer And Data API

## Build

```bash
# Option 1: Build to bin folder
go build -o bin/solution1

# Option 2: Install to GOPATH/bin (invoke as 'delegated')
go install
```

## Setup

```bash
# Set database connection URL
export DB_URL="postgresql://localhost/solution1"

# Create database
createdb solution1

# Load schema
psql solution1 < schema.sql
```

## Verify

```bash
# Connect to database
psql solution1

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
```

