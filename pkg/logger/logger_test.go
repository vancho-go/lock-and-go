package logger

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{
			name:     "Valid level - info",
			logLevel: "info",
			wantErr:  false,
		},
		{
			name:     "Valid level - error",
			logLevel: "error",
			wantErr:  false,
		},
		{
			name:     "Invalid level",
			logLevel: "notALevel",
			wantErr:  true,
		},
		{
			name:     "Empty level defaults to info",
			logLevel: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.logLevel)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}
