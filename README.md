# Solution 1: Simple Tezos Delegation Indexer And Data API

## Requirements

- Go 1.23 or above
- PostgreSQL

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
\q
```

## Usage

### Start Indexer

```bash
# Terminal 1: Start indexing delegations (runs every 60s)
export DB_URL="postgresql://localhost/solution1"
./bin/solution1 index
# OR if installed via go install:
delegated index
```

### Start API Server

```bash
# Terminal 2: Start the API server
export DB_URL="postgresql://localhost/solution1"
./bin/solution1 serve
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

