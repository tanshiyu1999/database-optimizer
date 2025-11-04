# Database Optimizer - Learning ModuleData Set From `https://www.kaggle.com/datasets/geraldooizx/g-coffee-shop-transaction-202307-to-202506`

A modular PostgreSQL database optimization learning project for students to explore indexing strategies and query performance profiling.

## 📁 Project Structure

```
database-optimizer/
├── main.go                  # Entry point - orchestrates the import process
├── schema.sql              # Database schema with table definition
├── .env                    # Environment variables (DATABASE_URL)
├── docker-compose.yml      # PostgreSQL container setup
├── data/
│   └── sf-fire-calls.csv  # San Francisco Fire Department calls dataset
├── db/
│   └── connection.go      # Database connection pool management
├── schema/
│   └── manager.go         # Schema operations and table statistics
├── importer/
│   └── csv_importer.go    # CSV data import with batch processing
└── profiler/
    └── profiler.go        # Performance profiling and metrics
```

## 🎯 Learning Objectives

1. **Understand Database Indexing**: Experiment with different index strategies
2. **Profile Query Performance**: Measure impact of indexes on query speed
3. **Optimize Batch Operations**: Learn efficient data import techniques
4. **Analyze Performance Metrics**: Interpret profiling data

## 🚀 Getting Started

### 1. Start PostgreSQL Database

```bash
docker-compose up -d
```

### 2. Import Data

```bash
go run main.go
```

This will:
- Create the `fire_calls` table
- Import 175,000+ records from CSV
- Display performance metrics
- Show table statistics

## 📊 Performance Metrics

The profiler tracks:
- **Schema creation time**: Time to create tables and indexes
- **Data import time**: Time to import all CSV records
- **Records per second**: Import throughput
- **Total execution time**: End-to-end performance

## 🔧 Module Overview

### `db/connection.go`
Manages PostgreSQL connection pooling using pgx/v5.

**Key Functions:**
- `Init()`: Initialize connection pool
- `GetPool()`: Access the connection pool
- `Close()`: Clean up resources

### `schema/manager.go`
Handles schema operations and statistics.

**Key Functions:**
- `CreateFromFile(filepath)`: Execute SQL schema file
- `GetTableStats()`: Retrieve table statistics

### `importer/csv_importer.go`
Imports CSV data using batch inserts for optimal performance.

**Key Functions:**
- `NewCSVImporter(pool, batchSize)`: Create importer
- `ImportFromFile(filepath)`: Import CSV data

**Students can experiment with:**
- Batch sizes (try 100, 1000, 5000)
- Different parsing strategies
- Error handling approaches

### `profiler/profiler.go`
Tracks and reports performance metrics.

**Key Functions:**
- `New()`: Create a new profiler
- `Start(operation)`: Begin timing an operation
- `End()`: Complete timing and store duration
- `PrintReport()`: Display formatted results

## 🎓 Student Exercises

### Exercise 1: Index Experimentation
1. Edit `schema.sql` to uncomment different indexes
2. Run the import with different index combinations
3. Compare import times with/without indexes
4. Document your findings

**Questions to explore:**
- How do indexes affect import speed?
- Which indexes provide the most query benefit?
- What's the cost vs. benefit tradeoff?

### Exercise 2: Batch Size Optimization
1. Modify batch size in `main.go` (line 37)
2. Test with: 100, 500, 1000, 2000, 5000 records per batch
3. Record "Records per second" for each
4. Find the optimal batch size

### Exercise 3: Query Performance
Add query profiling to measure:
- Queries with/without indexes
- Simple vs. complex WHERE clauses
- JOIN performance
- Aggregation queries

**Example query to profile:**
```go
// Add to main.go after import
queryOp := prof.Start("sample_query")
// Run your query here
queryOp.End()
```

### Exercise 4: Custom Indexes
Design and test custom indexes for common queries:
- Find all "Medical Incident" calls
- Count calls by neighborhood
- Filter by date range and call type
- Find high-priority calls

## 📈 Sample Queries to Test

```sql
-- Query 1: Count by call type (test idx_call_type)
SELECT call_type, COUNT(*) 
FROM fire_calls 
GROUP BY call_type 
ORDER BY COUNT(*) DESC;

-- Query 2: Filter by neighborhood (test idx_neighborhood)
SELECT * 
FROM fire_calls 
WHERE neighborhood = 'Mission' 
AND call_type = 'Medical Incident';

-- Query 3: Date range query (test idx_call_date)
SELECT COUNT(*) 
FROM fire_calls 
WHERE call_date BETWEEN '01/01/2002' AND '12/31/2002';

-- Query 4: Complex filter (test composite indexes)
SELECT unit_type, COUNT(*) 
FROM fire_calls 
WHERE call_type = 'Structure Fire' 
AND priority = 3 
GROUP BY unit_type;
```

## 🔍 Using EXPLAIN ANALYZE

Test query performance with PostgreSQL's EXPLAIN:

```sql
EXPLAIN ANALYZE 
SELECT * FROM fire_calls 
WHERE call_type = 'Medical Incident';
```

Look for:
- **Seq Scan**: Table scan (slow without index)
- **Index Scan**: Using an index (fast)
- **Execution Time**: Total query time

## 🛠️ Modifying for Your Needs

### Add New Profiling Metrics
```go
// In main.go
customOp := prof.Start("my_operation")
// ... your code ...
duration := customOp.End()
fmt.Printf("My operation took: %v\n", duration)
```

### Add New Statistics
```go
// In schema/manager.go - add to GetTableStats()
// Example: Count by unit type
rows, _ := m.pool.Query(ctx, `
    SELECT unit_type, COUNT(*) 
    FROM fire_calls 
    GROUP BY unit_type 
    ORDER BY COUNT(*) DESC 
    LIMIT 5
`)
```

### Test Different Import Strategies
Modify `importer/csv_importer.go` to try:
- Transaction batching
- COPY command instead of INSERT
- Parallel imports
- Pre-sorting data

## 📚 Additional Resources

- [PostgreSQL Performance Tips](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Understanding EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html)
- [Index Types in PostgreSQL](https://www.postgresql.org/docs/current/indexes-types.html)

## 🎯 Learning Goals Checklist

- [ ] Successfully import data and measure baseline performance
- [ ] Create at least 3 different index configurations
- [ ] Profile query performance with EXPLAIN ANALYZE
- [ ] Find optimal batch size for imports
- [ ] Document performance differences
- [ ] Create a composite index for complex queries
- [ ] Write custom profiling for queries
- [ ] Analyze the cost/benefit of different indexes

## 💡 Tips

1. **Always baseline first**: Run without indexes to establish baseline
2. **One change at a time**: Test one index/optimization at a time
3. **Real queries matter**: Index based on actual query patterns
4. **Document everything**: Keep notes on what works and what doesn't
5. **Reset between tests**: Use `docker-compose down -v` to start fresh

## 🐛 Troubleshooting

**Import fails with "numeric field overflow":**
- Check `schema.sql` - delay field should be `NUMERIC(15, 8)`

**Connection refused:**
- Ensure PostgreSQL is running: `docker-compose up -d`
- Check `.env` file has correct DATABASE_URL

**Slow imports:**
- Try larger batch sizes
- Disable indexes during import, add them after
- Check Docker container resources

---

**Happy Learning! 🚀**

For questions or issues, review the code comments or consult PostgreSQL documentation.
