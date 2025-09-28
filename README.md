# Auction Auto-Close Testing

## Run the app

```bash
docker-compose up --build
```

## Test steps

1. Create auction:

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{"product_name": "iPhone", "category": "Electronics", "description": "New iPhone", "condition": 1}'
```

2. Check active auctions:

```bash
curl "http://localhost:8080/auction?status=0"
```

3. Wait 20 seconds

4. Check again - auction should be closed:

```bash
curl "http://localhost:8080/auction?status=1"
```

## Faster testing

Change `AUCTION_INTERVAL=5s` in `cmd/auction/.env` and restart.
