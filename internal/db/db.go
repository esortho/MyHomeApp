package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Measurement struct {
	ID        int64
	Name      string
	Value     float64
	Unit      string
	Timestamp time.Time
}

func Init() error {
	var err error
	DB, err = sql.Open("sqlite3", "./measurements.db")
	if err != nil {
		return err
	}

	// Create measurements table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS measurements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value REAL NOT NULL,
			unit TEXT NOT NULL,
			timestamp DATETIME NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func StoreMeasurement(name string, value float64, unit string) error {
	_, err := DB.Exec(
		"INSERT INTO measurements (name, value, unit, timestamp) VALUES (?, ?, ?, ?)",
		name, value, unit, time.Now(),
	)
	return err
}

func GetHistoricalMeasurements(name string, from, to time.Time) ([]Measurement, error) {
	rows, err := DB.Query(
		"SELECT id, name, value, unit, timestamp FROM measurements WHERE name = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp",
		name, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measurements []Measurement
	for rows.Next() {
		var m Measurement
		err := rows.Scan(&m.ID, &m.Name, &m.Value, &m.Unit, &m.Timestamp)
		if err != nil {
			return nil, err
		}
		measurements = append(measurements, m)
	}

	return measurements, nil
}

func GetLatestMeasurements() (map[string]Measurement, error) {
	rows, err := DB.Query(`
		SELECT m1.id, m1.name, m1.value, m1.unit, m1.timestamp
		FROM measurements m1
		INNER JOIN (
			SELECT name, MAX(timestamp) as max_timestamp
			FROM measurements
			GROUP BY name
		) m2 ON m1.name = m2.name AND m1.timestamp = m2.max_timestamp
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	measurements := make(map[string]Measurement)
	for rows.Next() {
		var m Measurement
		err := rows.Scan(&m.ID, &m.Name, &m.Value, &m.Unit, &m.Timestamp)
		if err != nil {
			return nil, err
		}
		measurements[m.Name] = m
	}

	return measurements, nil
}

// GetMeasurementHistory retrieves historical measurements of a specific type within a time range
func GetMeasurementHistory(measurementType string, from, to time.Time) ([]Measurement, error) {
	var measurements []Measurement
	rows, err := DB.Query(
		"SELECT id, type, value, unit, updated_at FROM measurements WHERE type = ? AND updated_at BETWEEN ? AND ? ORDER BY updated_at ASC",
		measurementType, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query measurements: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Measurement
		err := rows.Scan(&m.ID, &m.Name, &m.Value, &m.Unit, &m.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan measurement: %v", err)
		}
		measurements = append(measurements, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating measurements: %v", err)
	}

	return measurements, nil
}
