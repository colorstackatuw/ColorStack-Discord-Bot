package database

import (
	"ColorStack-Discord-Bot/internal/types"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// this function tests if the same instance is being returned which should happen if connection pooling works correctly
func TestConnectionPooling(t *testing.T) {
	db1 = GetDatabaseInstance()
	db2 = GetDatabaseInstance()

	if (db1 != db2) {
		t.Errorf("The instance of the databases are not the same, pooling does not work")
	}
}	