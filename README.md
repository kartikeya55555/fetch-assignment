# fetch-assignment

A solution to Receipt Processor with in-memory storage.

## How to Run (Local)

1. **Run docker**:
   ```bash
   make run
   ```
2. **Integration tests**: 
   ```bash
   fetch-assignment % go test ./... -v 
            or 
   make test
   ```
3. **Manual testing**:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
   -d '{
        "retailer": "Apple Store",
        "purchaseDate": "2024-01-01",
        "purchaseTime": "14:01",
        "total": "12.00",
        "items": [
          { "shortDescription": "MacBook Charger", "price": "12.00" }
        ]
   }' http://localhost:8080/receipts/process
   ```
   Then retrieve points:
   ```bash
   curl http://localhost:8080/receipts/<returned-id>/points
   ```

## Folder structure 
```
FETCH-ASSIGNMENT/
├── api/
│   └── openapi/
│       └── api.yml                # OpenAPI spec for the receipt processor API
├── cmd/
│   └── main.go                    # Entry point: runs both API + worker in one process
├── internal/
│   ├── aws/
│   │   └── sqs_sns.go            # SQS and SNS client logic (AWS or LocalStack)
│   ├── config/
│   │   └── config.go             # Configuration loading from environment, etc.
│   ├── errors/
│   │   └── custom_errors.go      # Centralized custom error definitions
│   ├── handlers/
│   │   ├── receipt_handler.go    # HTTP handlers for receipts (POST / GET)
│   │   └── receipt_handler_test.go # Tests for these handlers (unit or integration)
│   ├── models/
│   │   ├── item.go               # Data model for an Item
│   │   └── receipt.go            # Data model for a Receipt
│   ├── receipt/
│   │   ├── points_calculator.go  # Logic for calculating points
│   │   ├── service.go            # Business logic for receipts (ProcessReceipt, etc.)
│   │   ├── store.go              # In-memory store for receipts
│   │   └── validate.go           # Validation logic for receipts
│   └── worker/
│       └── processor.go          # Worker code that processes SQS messages
├── .dockerignore
├── docker-compose.yml            # Docker Compose config (LocalStack + single container)
├── Dockerfile                    # Dockerfile for building the Go application
├── go.mod
├── go.sum
├── Makefile                      # Make targets for building, running, testing
└── README.md                     # This documentation
```