package logging

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	once   sync.Once
	logger *zerolog.Logger

	// Inject via CDK configuration
	LogLevel string = "0" // zerolog.Level: Trace = -1, Debug = 0, Info = 1, Error = 3, Disabled = 7
	CommitID string = "unknown"
)

func LogFunction(function string, start time.Time, msg string, state map[string]interface{}) {
	if logger == nil {
		_ = GetLogger() // Initialize logger but don't worry about the return since it's tied to this module
	}

	logger.Info().
		Str("internal_function", function).
		Int64("start_timestamp", start.Unix()).
		Int64("duration_ms", time.Now().Sub(start).Milliseconds()).
		Fields(state).
		Msg(msg)
}

func GetLogger() *zerolog.Logger {
	return GetLoggerWithContext(context.Background())
}

func GetLoggerWithContext(ctx context.Context) *zerolog.Logger {
	if logger == nil {
		once.Do(
			func() {
				zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
				lg := log.Logger

				// Read LogLevel from linker value
				var level zerolog.Level
				convertedLevel, err := strconv.ParseInt(LogLevel, 0, 8)
				if err != nil {
					level = zerolog.InfoLevel
					lg.Warn().Str("LogLevel", LogLevel).Msg("Unable to convert LogLevel to zerolog.Level")
				} else {
					level = zerolog.Level(convertedLevel)
				}

				// If possible, enhance the logger with info from the lambda context
				lc, ok := lambdacontext.FromContext(ctx)
				if ok && lc != nil {
					lg = log.With().
						Str("commit_id", CommitID).
						Str("request_id", lc.AwsRequestID).
						Str("lambda_function", lambdacontext.FunctionName).
						Logger().
						Level(zerolog.Level(level))
				} else {
					lg = log.With().
						Str("commit_id", CommitID).
						Logger().
						Level(zerolog.Level(level))
					lg.Warn().Msg("Lambda context not found")
				}

				// zerolog usage note: must use Msg() or Send() to trigger logs to actually send
				logger = &lg
				logger.Trace().Str("log_level", logger.GetLevel().String()).Msg("Logging initialized")
			},
		)
	}

	return logger
}
