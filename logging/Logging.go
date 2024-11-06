package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log zerolog.Logger

// Declare a logging system with zerolog
func init() {
	// Create the logger and location
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	lj := &lumberjack.Logger{
		Filename:   "logging/discord_bot.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	// Create a new zerolog logger with the lumberjack logger as the output
	log = zerolog.New(lj).With().Timestamp().Logger()
}

// Info logging
func Info(msg string) {
	log.Info().Msg(msg)
}

// Error logging
func Error(msg string, err error) {
	log.Error().Stack().Err(err).Msg(msg)
}

// Fatal loggin
func Fatal(msg string, err error) {
	log.Fatal().Stack().Err(err).Msg(msg)
}
