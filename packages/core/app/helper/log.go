package helper

import (
	"encoding/json"
	"log"
	"os"

	"github.com/rs/zerolog"
)

func LogJson(claims interface{}, name string) {
	claimsJSON, err := json.MarshalIndent(claims, "", "  ")
	if err != nil {
		log.Println("Error marshaling claims to JSON:", err)
	} else {
		log.Printf("%s: %s", name, claimsJSON)
	}
}

type logs struct {
	logger zerolog.Logger
	class  string
	name   string
}

func LogContext(class string, name string) logs {
	// Initialize zerolog with timestamp and service name
	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return logs{
		logger: baseLogger,
		class:  class,
		name:   name,
	}
}

func (l logs) Logger(value interface{}, name string, _isDebug ...bool) {
	isDebug := false

	if len(_isDebug) > 0 {
		isDebug = _isDebug[0]
	}

	serviceLogger := l.logger.With().Str(l.class, l.name).Logger()

	claimsJSON, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Println("Error marshaling value to JSON:", err)
	}

	serviceLogger = serviceLogger.With().RawJSON(name, claimsJSON).Logger()

	if isDebug {
		serviceLogger.Debug().Msg("Processed log debug values")
	} else {
		serviceLogger.Info().Msg("Processed log info values")
	}

}
