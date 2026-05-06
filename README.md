# Chip Transfer Exercise

This is a minimal in-memory peer-to-peer chip transfer backend with:

- API layer (`api`)
- Service layer (`service`)
- Concurrency-safe, atomic transfers

## Project Structure

```text
exercise/chip-transfer
├── main.go
├── README.md
├── api
│   ├── handler.go
│   └── handler_test.go
└── service
    ├── chip_service.go
    └── chip_service_test.go
```

## Run Locally

From repository root:

```bash
go run ./exercise/chip-transfer
```

Server starts on `:8080`.

## API Endpoints

### Transfer Chips

```bash
curl -X POST http://localhost:8080/transfer-chips \
  -H "Content-Type: application/json" \
  -d '{"fromPlayerId":"player-123","toPlayerId":"player-456","amount":2000}'
```

### Get Balance

```bash
curl http://localhost:8080/chip-balance/player-123
```

## Test

From repository root:

```bash
go test ./exercise/chip-transfer/...
```

## Step-by-Step Implementation

1. Define business rules in service constants and errors.
2. Implement in-memory balance store using `map[string]int64`.
3. Add a service-level mutex to protect all reads/writes.
4. Lazily initialize each unseen player with `10,000` chips.
5. Implement transfer validation:
    - sender and receiver IDs must be present
    - no self-transfer
    - amount must be `> 0` and `<= 5,000`
    - sender must have enough chips
6. Make transfer atomic by running debit and credit in one lock scope.
7. Build HTTP handlers:
    - `POST /transfer-chips`
    - `GET /chip-balance/{playerId}`
8. Return clear JSON errors with suitable status codes.
9. Add unit tests for rules and a concurrency test for race-safe totals.
