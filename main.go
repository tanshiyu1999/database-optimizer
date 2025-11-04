package main

import (
	"database-optimizer/db"
	"database-optimizer/importer"
	"database-optimizer/profiler"
	"database-optimizer/schema"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	// Initialize database connection
	db.Init()
	defer db.Close()

	fmt.Println("🚀 Starting data import...")

	// Create profiler to track performance
	prof := profiler.New()

	// Step 1: Create schema
	schemaOp := prof.Start("schema_creation")
	schemaManager := schema.NewManager(db.GetPool())
	if err := schemaManager.CreateFromFile("schema.sql"); err != nil {
		log.Fatalf("❌ Failed to create schema: %v", err)
	}
	schemaDuration := schemaOp.End()
	fmt.Printf("✅ Schema created in: %v\n", schemaDuration)

	// Step 2: Import CSV data
	importOp := prof.Start("data_import")
	csvImporter := importer.NewCSVImporter(db.GetPool(), 1000) // 1000 records per batch

	// Choose import method:
	// Option 1: Sequential import (simpler, easier to debug)
	// recordsImported, err := csvImporter.ImportFromFile("data/sf-fire-calls.csv")

	// Option 2: Concurrent import with goroutines (faster for large datasets)
	recordsImported, err := csvImporter.ImportFromFileGoRoutine("data/sf-fire-calls.csv", 4) // 4 workers

	if err != nil {
		log.Fatalf("❌ Failed to import data: %v", err)
	}
	importOp.End()

	// Step 3: Print profiling report
	prof.PrintReport(recordsImported)

	// Step 4: Get and display table statistics
	stats, err := schemaManager.GetTableStats()
	if err != nil {
		log.Printf("⚠️  Warning: Could not retrieve table stats: %v", err)
	} else {
		profiler.PrintTableStats(stats.TotalRecords, convertCallTypes(stats.TopCallTypes))
	}
}

// Helper function to convert call types for printing
func convertCallTypes(callTypes []schema.CallTypeCount) []struct {
	CallType string
	Count    int
} {
	result := make([]struct {
		CallType string
		Count    int
	}, len(callTypes))
	for i, ct := range callTypes {
		result[i].CallType = ct.CallType
		result[i].Count = ct.Count
	}
	return result
}
