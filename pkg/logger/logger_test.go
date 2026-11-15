package logger

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestNewLoggerLevel(t *testing.T) {
	log := NewLogger(zerolog.WarnLevel, false)
	if log.GetLevel() != zerolog.WarnLevel {
		t.Fatalf("level = %s", log.GetLevel())
	}

	pretty := NewLogger(zerolog.ErrorLevel, true)
	if pretty.GetLevel() != zerolog.ErrorLevel {
		t.Fatalf("pretty level = %s", pretty.GetLevel())
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  zerolog.Level
	}{
		{name: "debug", level: "debug", want: zerolog.DebugLevel},
		{name: "info", level: "info", want: zerolog.InfoLevel},
		{name: "error", level: "error", want: zerolog.ErrorLevel},
		{name: "unknown falls back to debug", level: "nope", want: zerolog.DebugLevel},
		{name: "empty is no level", level: "", want: zerolog.NoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New(tt.level, false)
			if log.GetLevel() != tt.want {
				t.Fatalf("level = %s, want %s", log.GetLevel(), tt.want)
			}
		})
	}
}
