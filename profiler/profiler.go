package profiler

import (
	"fmt"
	"time"
)

// Profiler tracks performance metrics for database operations
type Profiler struct {
	operations map[string]time.Duration
	startTimes map[string]time.Time
}

// Operation represents a timed operation
type Operation struct {
	name     string
	profiler *Profiler
}

// New creates a new profiler
func New() *Profiler {
	return &Profiler{
		operations: make(map[string]time.Duration),
		startTimes: make(map[string]time.Time),
	}
}

// Start begins timing an operation
func (p *Profiler) Start(operationName string) *Operation {
	p.startTimes[operationName] = time.Now()
	return &Operation{
		name:     operationName,
		profiler: p,
	}
}

// End completes the timing of an operation and returns the duration
func (op *Operation) End() time.Duration {
	startTime, exists := op.profiler.startTimes[op.name]
	if !exists {
		return 0
	}

	duration := time.Since(startTime)
	op.profiler.operations[op.name] = duration
	delete(op.profiler.startTimes, op.name)

	return duration
}

// GetDuration returns the duration of a completed operation
func (p *Profiler) GetDuration(operationName string) time.Duration {
	return p.operations[operationName]
}

// PrintReport prints a formatted performance report
func (p *Profiler) PrintReport(totalRecords int) {
	fmt.Println("\n" + strings("=", 60))
	fmt.Println("📊 PERFORMANCE REPORT")
	fmt.Println(strings("=", 60))

	// Calculate total time
	var totalTime time.Duration
	for _, duration := range p.operations {
		totalTime += duration
	}

	// Print individual operations
	for name, duration := range p.operations {
		fmt.Printf("⏱️  %-20s: %v\n", name, duration)
	}

	fmt.Println(strings("-", 60))

	// Calculate and display throughput
	if totalRecords > 0 {
		importDuration := p.operations["data_import"]
		if importDuration > 0 {
			recordsPerSecond := float64(totalRecords) / importDuration.Seconds()
			fmt.Printf("📈 Records imported    : %d\n", totalRecords)
			fmt.Printf("⚡ Records per second  : %.2f\n", recordsPerSecond)
			fmt.Println(strings("-", 60))
		}
	}

	fmt.Printf("🏁 Total execution time: %v\n", totalTime)
	fmt.Println(strings("=", 60) + "\n")
}

// PrintTableStats prints statistics about the database table
func PrintTableStats(totalRecords int, topCallTypes []struct {
	CallType string
	Count    int
}) {
	fmt.Println("\n" + strings("=", 60))
	fmt.Println("📋 TABLE STATISTICS")
	fmt.Println(strings("=", 60))
	fmt.Printf("📊 Total records: %d\n", totalRecords)
	fmt.Println("\n🔥 Top 5 Call Types:")
	fmt.Println(strings("-", 60))

	for i, ct := range topCallTypes {
		percentage := float64(ct.Count) / float64(totalRecords) * 100
		fmt.Printf("%d. %-30s: %6d (%.2f%%)\n", i+1, ct.CallType, ct.Count, percentage)
	}

	fmt.Println(strings("=", 60) + "\n")
}

// strings creates a string by repeating a character n times
func strings(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}

// Reset clears all profiling data
func (p *Profiler) Reset() {
	p.operations = make(map[string]time.Duration)
	p.startTimes = make(map[string]time.Time)
}

// GetAllOperations returns a copy of all recorded operations
func (p *Profiler) GetAllOperations() map[string]time.Duration {
	operations := make(map[string]time.Duration)
	for k, v := range p.operations {
		operations[k] = v
	}
	return operations
}

// FormatDuration formats a duration into a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Microseconds()))
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Milliseconds()))
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else {
		minutes := int(d.Minutes())
		seconds := d.Seconds() - float64(minutes*60)
		return fmt.Sprintf("%dm %.2fs", minutes, seconds)
	}
}

// CompareResults compares two profiling results and prints the difference
func CompareResults(name1 string, duration1 time.Duration, name2 string, duration2 time.Duration) {
	fmt.Println("\n" + strings("=", 60))
	fmt.Println("⚖️  PERFORMANCE COMPARISON")
	fmt.Println(strings("=", 60))
	fmt.Printf("%-30s: %v\n", name1, duration1)
	fmt.Printf("%-30s: %v\n", name2, duration2)
	fmt.Println(strings("-", 60))

	if duration1 > duration2 {
		improvement := float64(duration1-duration2) / float64(duration1) * 100
		speedup := float64(duration1) / float64(duration2)
		fmt.Printf("✅ %s is %.2f%% faster (%.2fx speedup)\n", name2, improvement, speedup)
	} else if duration2 > duration1 {
		improvement := float64(duration2-duration1) / float64(duration2) * 100
		speedup := float64(duration2) / float64(duration1)
		fmt.Printf("✅ %s is %.2f%% faster (%.2fx speedup)\n", name1, improvement, speedup)
	} else {
		fmt.Println("⚖️  Both methods took the same time")
	}

	fmt.Println(strings("=", 60) + "\n")
}
