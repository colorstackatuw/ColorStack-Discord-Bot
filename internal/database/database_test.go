package database

import (
	"ColorStack-Discord-Bot/internal/types"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// CreateTestSQLiteDB sets up an in-memory SQLite DB for testing
func createTestSQLiteDB(t *testing.T) ChannelsDB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	// Drop the table if it exists first
	_, err = db.Exec(`DROP TABLE IF EXISTS dim_server`)
	if err != nil {
		t.Fatalf("failed to drop existing table: %v", err)
	}

	// Create a compatible schema
	_, err = db.Exec(`
		CREATE TABLE dim_server (
			server_id TEXT,
			join_date TIMESTAMP,
			server_name TEXT,
			channel_id INTEGER,
			bot_type TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return &Database{conn: db}
}

func TestWriteChannel(t *testing.T) {
	db := createTestSQLiteDB(t)

	err := db.WriteChannel("server-1", "Guild Test", "1001", types.SLACK)
	if err != nil {
		t.Fatalf("WriteChannel failed: %v", err)
	}

	// Validate insertion
	rows, err := db.(*Database).conn.Query(
		"SELECT COUNT(*) FROM dim_server WHERE server_id = ?",
		"server-1",
	)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
	}

	if count != 1 {
		t.Fatalf("Expected 1 row, got %d", count)
	}
}

func TestGetChannels(t *testing.T) {
	db := createTestSQLiteDB(t)

	// Insert test data
	_ = db.WriteChannel("server-2", "Guild Test", "2002", types.DISCORD)
	_ = db.WriteChannel("server-2", "Guild Test", "2003", types.DISCORD)

	channels, err := db.GetChannels()
	if err != nil {
		t.Fatalf("GetChannels failed: %v", err)
	}

	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}
}

func TestDeleteServer(t *testing.T) {
	db := createTestSQLiteDB(t)

	_ = db.WriteChannel("server-3", "Guild Test", "3001", types.DISCORD)

	err := db.DeleteServer("server-3")
	if err != nil {
		t.Fatalf("DeleteServer failed: %v", err)
	}

	// Ensure the row is deleted
	rows, err := db.(*Database).conn.Query(
		"SELECT COUNT(*) FROM dim_server WHERE server_id = ?",
		"server-3",
	)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
	}

	if count != 0 {
		t.Errorf("Expected 0 rows, got %d", count)
	}
}
