package database

import (
	log "ColorStack-Discord-Bot/internal/logger"
	"ColorStack-Discord-Bot/internal/types"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/godror/godror"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var (
	once     sync.Once
	instance ChannelsDB
)

type ChannelsDB interface {
	WriteChannel(guildID string, guildName string, channelID string, botType types.BotType) error
	GetChannels() ([]int64, error)
	DeleteServer(guildID string) error
	Close() error
}

// DatabaseService manages the connection and operations with Oracle DB.
type Database struct {
	conn *sql.DB
}

// Returns single version of singleton database
func GetDatabaseInstance() ChannelsDB {
	once.Do(func() {
		service, err := newDatabaseService()
		if err != nil {
			log.Fatal("Failed to initialize database service: %v", err)
		}
		instance = service
	})
	return instance
}

// NewDatabaseService initializes a new DatabaseService instance.
func newDatabaseService() (ChannelsDB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment variables: %v", err)
	}

	// Retrieve Oracle DB credentials from environment variables
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dsn := os.Getenv("DB_DSN")

	if dsn == "" || username == "" || password == "" {
		return nil, errors.Wrap(err, "One of following returned empty! DSN, Password, Username")
	}

	// Create Oracle DB connection
	connString := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, username, password, dsn)
	conn, err := sql.Open("godror", connString)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't connect to oracle database!")
	}

	return &Database{conn: conn}, nil
}

// WriteChannel inserts the channel data into the Oracle database.
func (db *Database) WriteChannel(
	guildID string,
	guildName string,
	channelID string,
	botType BotType,
) error {
	query := `
		INSERT INTO dim_server 
		(server_id, join_date, server_name, channel_id, bot_type) 
		VALUES (:1, :2, :3, :4, :5)
	`
	_, err := db.conn.Exec(query, guildID, time.Now(), guildName, channelID, string(botType))
	if err != nil {
		msg := fmt.Sprintf(
			"Failed to insert channel into database! GuildID: %s, ChannelID: %s",
			guildID,
			channelID,
		)
		log.Error(msg, err)
		return errors.Wrap(err, "Couldn't write channel to db")
	}

	log.Info("Collected channel!")
	return nil
}

// GetChannels retrieves all unique channel IDs from the Oracle database.
func (db *Database) GetChannels() ([]int64, error) {
	query := "SELECT DISTINCT channel_id FROM dim_server"

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channelIDs []int64
	for rows.Next() {
		var channelID int64
		if err := rows.Scan(&channelID); err != nil {
			errors.Wrap(err, "Error scnning channel ID")
			return nil, err
		}
		channelIDs = append(channelIDs, channelID)
	}

	msg := fmt.Sprintf("Successfully collect %d channels!", len(channelIDs))
	log.Info(msg)
	return channelIDs, nil
}

// DeleteServer removes all records associated with a specific server from the Oracle database.
func (db *Database) DeleteServer(guildID string) error {
	query := "DELETE FROM dim_server WHERE server_id = :1"

	_, err := db.conn.Exec(query, guildID)
	if err != nil {
		msg := fmt.Sprintf("Failed to insert channel into database! GuildID: %s", guildID)
		log.Error(msg, err)
		errors.Wrap(err, "Failed to delete server!")
	}

	log.Info("Successfully deleted channel")
	return nil
}

// Close connection to database
func (d *Database) Close() error {
	return d.conn.Close()
}

// Assert Database implements ChannelsDB interface
var _ ChannelsDB = (*Database)(nil)
