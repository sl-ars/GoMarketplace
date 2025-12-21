# Elasticsearch Re-indexing Script

This script re-indexes all products with their current offers in Elasticsearch. Useful when:

- Products exist in database but not in Elasticsearch
- You want to refresh all product search data
- Offers have been updated and need re-indexing

## Usage

```bash
# Make sure Elasticsearch is running
docker-compose up -d elasticsearch

# Run the re-indexing script
go run ./cmd/reindex_es --config ./configs/.env

# Or build and run
go build ./cmd/reindex_es
./reindex_es --config ./configs/.env
```

## What it does

1. **Connects** to both PostgreSQL and Elasticsearch
2. **Fetches all products** from the database
3. **For each product**, gets all current offers
4. **Re-indexes** each product with pricing data from offers
5. **Reports** success/failure counts

## Output Example

```
🔄 Starting Elasticsearch re-indexing...
Found 20 products to index
  ✓ Indexed product: iPhone 17 Pro Max (ID: 1, 3 offers)
  ✓ Indexed product: Samsung Galaxy S25 Ultra (ID: 2, 2 offers)
  ...

✅ Successfully indexed 20/20 products in Elasticsearch!
📊 To verify: curl localhost:9200/products/_count
```

## Verification

After running, verify the index contains documents:

```bash
# Check document count
curl -X GET "localhost:9200/products/_count?pretty"

# Search for products
curl -X GET "localhost:9200/products/_search?q=iphone&pretty"

# Your API should now return search results
curl "http://localhost:8080/api/products/search?q=iphone"
```

## When to use

- **After seeding data**: `go run ./cmd/seed_data` then `go run ./cmd/reindex_es`
- **After bulk updates**: When many offers change at once
- **Data sync issues**: When Elasticsearch and database are out of sync
- **Fresh deployment**: To populate search index with existing data

## Notes

- **Safe to run multiple times** - uses upsert operations
- **Asynchronous** - won't block your application
- **Reports progress** - shows which products are being indexed
- **Requires ES connection** - fails gracefully if ES is unavailable
