package schema

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Manager handles database schema operations
type Manager struct {
	pool *pgxpool.Pool
}

// CallTypeCount represents a call type and its count
type CallTypeCount struct {
	CallType string
	Count    int
}

// TableStats holds statistics about the fire_calls table
type TableStats struct {
	TotalRecords int
	TopCallTypes []CallTypeCount
}

// NewManager creates a new schema manager
func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{pool: pool}
}

// CreateFromFile reads and executes a SQL schema file
func (m *Manager) CreateFromFile(filepath string) error {
	ctx := context.Background()

	// Read the SQL file
	sqlBytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute the SQL commands
	_, err = m.pool.Exec(ctx, string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// GetTableStats retrieves statistics about the fire_calls table
func (m *Manager) GetTableStats() (*TableStats, error) {
	ctx := context.Background()
	stats := &TableStats{}

	// Get total record count
	err := m.pool.QueryRow(ctx, "SELECT COUNT(*) FROM fire_calls").Scan(&stats.TotalRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to get total records: %w", err)
	}

	// Get top 5 call types
	rows, err := m.pool.Query(ctx, `
		SELECT call_type, COUNT(*) as count 
		FROM fire_calls 
		GROUP BY call_type 
		ORDER BY count DESC 
		LIMIT 5
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get top call types: %w", err)
	}
	defer rows.Close()

	stats.TopCallTypes = make([]CallTypeCount, 0, 5)
	for rows.Next() {
		var ct CallTypeCount
		if err := rows.Scan(&ct.CallType, &ct.Count); err != nil {
			return nil, fmt.Errorf("failed to scan call type: %w", err)
		}
		stats.TopCallTypes = append(stats.TopCallTypes, ct)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return stats, nil
}

// DropTable drops the fire_calls table if it exists
func (m *Manager) DropTable() error {
	ctx := context.Background()
	_, err := m.pool.Exec(ctx, "DROP TABLE IF EXISTS fire_calls")
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}
	return nil
}

// CreateIndexes creates all recommended indexes for the fire_calls table
func (m *Manager) CreateIndexes() error {
	ctx := context.Background()

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_call_type ON fire_calls(call_type)",
		"CREATE INDEX IF NOT EXISTS idx_call_date ON fire_calls(call_date)",
		"CREATE INDEX IF NOT EXISTS idx_neighborhood ON fire_calls(neighborhood)",
		"CREATE INDEX IF NOT EXISTS idx_unit_type ON fire_calls(unit_type)",
	}

	for _, indexSQL := range indexes {
		if _, err := m.pool.Exec(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// DropIndexes drops all indexes from the fire_calls table
func (m *Manager) DropIndexes() error {
	ctx := context.Background()

	indexes := []string{
		"DROP INDEX IF EXISTS idx_call_type",
		"DROP INDEX IF EXISTS idx_call_date",
		"DROP INDEX IF EXISTS idx_neighborhood",
		"DROP INDEX IF EXISTS idx_unit_type",
	}

	for _, indexSQL := range indexes {
		if _, err := m.pool.Exec(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to drop index: %w", err)
		}
	}

	return nil
}
