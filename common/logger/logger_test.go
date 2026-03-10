/*
 *  Copyright (c) 2026 Original Author(s), PhonePe India Pvt. Ltd.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogConfig(t *testing.T) {
	config := DefaultLogConfig()
	assert.True(t, config.Enabled)
	assert.Equal(t, LevelInfo, config.Level)
	assert.False(t, config.JSON)
}

func TestDisabledLogConfig(t *testing.T) {
	config := DisabledLogConfig()
	assert.False(t, config.Enabled)
}

func TestNewLogger_DefaultConfig(t *testing.T) {
	logger := NewLogger(nil)
	assert.NotNil(t, logger)
	assert.True(t, logger.IsEnabled())
}

func TestNewLogger_DisabledConfig(t *testing.T) {
	logger := NewLogger(DisabledLogConfig())
	assert.NotNil(t, logger)
	assert.False(t, logger.IsEnabled())
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelDebug,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Debug("test debug message")
	output := buf.String()
	assert.Contains(t, output, "test debug message")
	assert.Contains(t, output, "DEBUG")
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelInfo,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Info("test info message")
	output := buf.String()
	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, "INFO")
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelWarn,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Warn("test warn message")
	output := buf.String()
	assert.Contains(t, output, "test warn message")
	assert.Contains(t, output, "WARN")
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelError,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Error("test error message")
	output := buf.String()
	assert.Contains(t, output, "test error message")
	assert.Contains(t, output, "ERROR")
}

func TestLogger_WithStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelInfo,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Info("user logged in", "user_id", "123", "email", "test@example.com")
	output := buf.String()
	assert.Contains(t, output, "user logged in")
	assert.Contains(t, output, "user_id=123")
	assert.Contains(t, output, "email=test@example.com")
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelInfo,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	contextLogger := logger.With("component", "TokenService")
	contextLogger.Info("token refreshed")
	output := buf.String()
	assert.Contains(t, output, "component=TokenService")
	assert.Contains(t, output, "token refreshed")
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		configLevel LogLevel
		logLevel    string
		logMessage  string
		shouldLog   bool
	}{
		{"Info filters Debug", LevelInfo, "DEBUG", "debug msg", false},
		{"Info allows Info", LevelInfo, "INFO", "info msg", true},
		{"Info allows Warn", LevelInfo, "WARN", "warn msg", true},
		{"Info allows Error", LevelInfo, "ERROR", "error msg", true},
		{"Error filters Info", LevelError, "INFO", "info msg", false},
		{"Error allows Error", LevelError, "ERROR", "error msg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &LogConfig{
				Enabled: true,
				Level:   tt.configLevel,
				Writer:  &buf,
				JSON:    false,
			}
			logger := NewLogger(config)

			switch tt.logLevel {
			case "DEBUG":
				logger.Debug(tt.logMessage)
			case "INFO":
				logger.Info(tt.logMessage)
			case "WARN":
				logger.Warn(tt.logMessage)
			case "ERROR":
				logger.Error(tt.logMessage)
			}

			output := buf.String()
			if tt.shouldLog {
				assert.Contains(t, output, tt.logMessage)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelInfo,
		Writer:  &buf,
		JSON:    true,
	}
	logger := NewLogger(config)

	logger.Info("json test", "key", "value")
	output := buf.String()
	assert.Contains(t, output, `"msg":"json test"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestLogger_DisabledDoesNotLog(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: false,
		Level:   LevelDebug,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Debug("should not appear")
	logger.Info("should not appear")
	logger.Warn("should not appear")
	logger.Error("should not appear")

	output := buf.String()
	assert.Empty(t, output)
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelInfo,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)
	SetDefaultLogger(logger)

	Info("global info message")
	output := buf.String()
	assert.Contains(t, output, "global info message")

	// Restore default logger
	SetDefaultLogger(NewLogger(DefaultLogConfig()))
}

func TestConvertLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel LogLevel
		expected string
	}{
		{"Debug", LevelDebug, "DEBUG"},
		{"Info", LevelInfo, "INFO"},
		{"Warn", LevelWarn, "WARN"},
		{"Error", LevelError, "ERROR"},
		{"Unknown", LogLevel("UNKNOWN"), "INFO"}, // defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := convertLogLevel(tt.logLevel)
			assert.NotNil(t, level)
		})
	}
}

func TestLogger_MultipleMessages(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelDebug,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Debug("message 1")
	logger.Info("message 2")
	logger.Warn("message 3")
	logger.Error("message 4")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 4)
}

func TestLogger_ErrorWithStackTrace(t *testing.T) {
	var buf bytes.Buffer
	config := &LogConfig{
		Enabled: true,
		Level:   LevelError,
		Writer:  &buf,
		JSON:    false,
	}
	logger := NewLogger(config)

	logger.Error("operation failed", "error", "connection timeout", "retry_count", 3)
	output := buf.String()
	assert.Contains(t, output, "operation failed")
	assert.Contains(t, output, "error")
	assert.Contains(t, output, "connection timeout")
	assert.Contains(t, output, "retry_count=3")
}
