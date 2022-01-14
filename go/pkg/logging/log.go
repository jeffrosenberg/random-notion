package logging

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO: Can this be a method on *zerolog.Logger, instead of a function that requires passing the logger?
// TODO: Clean up method signature
// TODO: Implement singleton pattern and decouple this from api interface?
func LogFunction(logger *zerolog.Logger, function string, start time.Time, msg string, state map[string]interface{}) {
	if logger == nil {
		logger = &log.Logger
	}

	logger.Info().
		Str("internal_function", function).
		Int64("start_timestamp", start.Unix()).
		Int64("duration_ms", time.Now().Sub(start).Milliseconds()).
		Fields(state).
		Msg(msg)
}
