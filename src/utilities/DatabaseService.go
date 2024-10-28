package utilities

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/godror/godror"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// DatabaseService manages the connection and operations with Oracle DB.
type DatabaseService struct {
	conn *sql.DB
}

// NewDatabaseService initializes a new DatabaseService instance.
func NewDatabaseService() (*DatabaseService, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
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

	return &DatabaseService{conn: conn}, nil
}

// WriteChannel inserts the channel data into the Oracle database.
func (db *DatabaseService) WriteChannel(guild *discordgo.Guild, channel *discordgo.Channel) error {
	query := `
		INSERT INTO DISCORD.DIM_DISCORD_TOKENS 
		(server_id, join_date, server_name, channel_id) 
		VALUES (:1, :2, :3, :4)
	`

	_, err := db.conn.Exec(query, guild.ID, time.Now(), guild.Name, channel.ID)
	if err != nil {
		return errors.Wrap(err, "Couldn't write channel to db")
	} else {
		return nil
	}
}

// GetChannels retrieves all unique channel IDs from the Oracle database.
func (db *DatabaseService) GetChannels() ([]int64, error) {
	query := "SELECT DISTINCT channel_id FROM DISCORD.DIM_DISCORD_TOKENS"
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
	return channelIDs, nil
}

// DeleteServer removes all records associated with a specific server from the Oracle database.
func (db *DatabaseService) DeleteServer(guild *discordgo.Guild) error {
	query := "DELETE FROM DISCORD.DIM_DISCORD_TOKENS WHERE server_id = :1"

	_, err := db.conn.Exec(query, guild.ID)
	if err != nil {
		errors.Wrap(err, "Failed to delete server!")
	}

	return nil
}
