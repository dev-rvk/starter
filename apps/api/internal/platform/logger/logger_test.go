package logger

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		level     string
		pretty    bool
		wantLevel zerolog.Level
	}{
		{
			name:      "info level pretty",
			level:     "info",
			pretty:    true,
			wantLevel: zerolog.InfoLevel,
		},
		{
			name:      "debug level JSON",
			level:     "debug",
			pretty:    false,
			wantLevel: zerolog.DebugLevel,
		},
		{
			name:      "invalid level fallback to info",
			level:     "invalid",
			pretty:    false,
			wantLevel: zerolog.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = New(tt.level, tt.pretty)

			if zerolog.GlobalLevel() != tt.wantLevel {
				t.Errorf("expected global level %s, got %s", tt.wantLevel, zerolog.GlobalLevel())
			}
		})
	}
}
