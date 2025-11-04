package importer

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CSVImporter handles importing CSV data into PostgreSQL
type CSVImporter struct {
	pool      *pgxpool.Pool
	batchSize int
}

// NewCSVImporter creates a new CSV importer with the specified batch size
func NewCSVImporter(pool *pgxpool.Pool, batchSize int) *CSVImporter {
	return &CSVImporter{
		pool:      pool,
		batchSize: batchSize,
	}
}

// ImportFromFile imports data from a CSV file sequentially (single-threaded)
func (imp *CSVImporter) ImportFromFile(filepath string) (int, error) {
	ctx := context.Background()

	// Open the CSV file
	file, err := os.Open(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read and skip header
	_, err = reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV header: %w", err)
	}

	totalRecords := 0
	batch := make([][]interface{}, 0, imp.batchSize)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return totalRecords, fmt.Errorf("error reading CSV record: %w", err)
		}

		// Parse the record
		row, err := imp.parseRow(record)
		if err != nil {
			// Skip invalid rows but log them
			fmt.Printf("⚠️  Skipping invalid row: %v\n", err)
			continue
		}

		batch = append(batch, row)

		// Insert batch when it reaches the batch size
		if len(batch) >= imp.batchSize {
			if err := imp.executeBatch(ctx, batch); err != nil {
				return totalRecords, fmt.Errorf("failed to insert batch: %w", err)
			}
			totalRecords += len(batch)
			batch = batch[:0] // Reset batch
		}
	}

	// Insert remaining records
	if len(batch) > 0 {
		if err := imp.executeBatch(ctx, batch); err != nil {
			return totalRecords, fmt.Errorf("failed to insert final batch: %w", err)
		}
		totalRecords += len(batch)
	}

	return totalRecords, nil
}

// ImportFromFileGoRoutine imports data from a CSV file using concurrent workers
func (imp *CSVImporter) ImportFromFileGoRoutine(filepath string, numWorkers int) (int, error) {
	ctx := context.Background()

	// Open the CSV file
	file, err := os.Open(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read and skip header
	_, err = reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Create channels for pipeline
	recordChan := make(chan []string, imp.batchSize*2)
	batchChan := make(chan [][]interface{}, numWorkers)
	errorChan := make(chan error, numWorkers)
	doneChan := make(chan bool)

	var totalRecords int
	var recordMutex sync.Mutex

	// Stage 1: CSV Reader goroutine
	go func() {
		defer close(recordChan)
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				errorChan <- fmt.Errorf("error reading CSV: %w", err)
				return
			}
			recordChan <- record
		}
	}()

	// Stage 2: Parse workers
	var parseWg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		parseWg.Add(1)
		go func(workerID int) {
			defer parseWg.Done()
			batch := make([][]interface{}, 0, imp.batchSize)

			for record := range recordChan {
				row, err := imp.parseRow(record)
				if err != nil {
					// Skip invalid rows
					continue
				}

				batch = append(batch, row)

				if len(batch) >= imp.batchSize {
					// Send batch to insert channel
					batchChan <- batch
					batch = make([][]interface{}, 0, imp.batchSize)
				}
			}

			// Send remaining records
			if len(batch) > 0 {
				batchChan <- batch
			}
		}(i)
	}

	// Close batchChan when all parse workers are done
	go func() {
		parseWg.Wait()
		close(batchChan)
	}()

	// Stage 3: Insert workers
	var insertWg sync.WaitGroup
	insertWorkers := 2 // Limit concurrent database connections

	for i := 0; i < insertWorkers; i++ {
		insertWg.Add(1)
		go func(workerID int) {
			defer insertWg.Done()

			for batch := range batchChan {
				if err := imp.executeBatch(ctx, batch); err != nil {
					errorChan <- fmt.Errorf("worker %d: %w", workerID, err)
					return
				}

				recordMutex.Lock()
				totalRecords += len(batch)
				recordMutex.Unlock()
			}
		}(i)
	}

	// Wait for all insert workers to complete
	go func() {
		insertWg.Wait()
		doneChan <- true
	}()

	// Wait for completion or error
	select {
	case err := <-errorChan:
		return totalRecords, err
	case <-doneChan:
		return totalRecords, nil
	}
}

// parseRow converts a CSV record into a database row
func (imp *CSVImporter) parseRow(record []string) ([]interface{}, error) {
	if len(record) != 28 {
		return nil, fmt.Errorf("expected 28 fields, got %d", len(record))
	}

	// Helper function to parse nullable integers
	parseInt := func(s string) interface{} {
		if s == "" {
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}
		return val
	}

	// Helper function to parse nullable floats
	parseFloat := func(s string) interface{} {
		if s == "" {
			return nil
		}
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil
		}
		return val
	}

	// Helper function to parse boolean
	parseBool := func(s string) interface{} {
		s = strings.ToLower(strings.TrimSpace(s))
		return s == "true" || s == "t" || s == "1"
	}

	// Helper function to handle nullable strings
	parseString := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}

	row := []interface{}{
		parseString(record[0]),  // call_number
		parseString(record[1]),  // unit_id
		parseInt(record[2]),     // incident_number
		parseString(record[3]),  // call_type
		parseString(record[4]),  // call_date
		parseString(record[5]),  // watch_date
		parseString(record[6]),  // call_final_disposition
		parseString(record[7]),  // available_dt_tm
		parseString(record[8]),  // address
		parseString(record[9]),  // city
		parseString(record[10]), // zipcode
		parseString(record[11]), // battalion
		parseString(record[12]), // station_area
		parseString(record[13]), // box
		parseInt(record[14]),    // original_priority
		parseInt(record[15]),    // priority
		parseInt(record[16]),    // final_priority
		parseBool(record[17]),   // als_unit
		parseString(record[18]), // call_type_group
		parseInt(record[19]),    // num_alarms
		parseString(record[20]), // unit_type
		parseInt(record[21]),    // unit_sequence_in_call_dispatch
		parseString(record[22]), // fire_prevention_district
		parseString(record[23]), // supervisor_district
		parseString(record[24]), // neighborhood
		parseString(record[25]), // location
		parseString(record[26]), // row_id (PRIMARY KEY)
		parseFloat(record[27]),  // delay
	}

	return row, nil
}

// executeBatch inserts a batch of rows into the database
func (imp *CSVImporter) executeBatch(ctx context.Context, batch [][]interface{}) error {
	if len(batch) == 0 {
		return nil
	}

	// Build the INSERT statement with placeholders
	query := `INSERT INTO fire_calls (
		call_number, unit_id, incident_number, call_type, call_date, watch_date,
		call_final_disposition, available_dt_tm, address, city, zipcode, battalion,
		station_area, box, original_priority, priority, final_priority, als_unit,
		call_type_group, num_alarms, unit_type, unit_sequence_in_call_dispatch,
		fire_prevention_district, supervisor_district, neighborhood, location, row_id, delay
	) VALUES `

	// Build value placeholders
	values := make([]interface{}, 0, len(batch)*28)
	placeholders := make([]string, 0, len(batch))

	for i, row := range batch {
		placeholder := "("
		for j := 0; j < 28; j++ {
			if j > 0 {
				placeholder += ","
			}
			placeholder += fmt.Sprintf("$%d", i*28+j+1)
			values = append(values, row[j])
		}
		placeholder += ")"
		placeholders = append(placeholders, placeholder)
	}

	query += strings.Join(placeholders, ",")

	// Execute the batch insert
	_, err := imp.pool.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	return nil
}
