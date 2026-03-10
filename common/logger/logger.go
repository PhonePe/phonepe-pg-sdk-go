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
	"io"
	"log/slog"
	"os"
)

// LogLevel represents the severity level of log messages
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogConfig holds the configuration for the SDK logger
type LogConfig struct {
	Enabled bool      // Enable or disable logging
	Level   LogLevel  // Minimum log level to output
	Writer  io.Writer // Output destination (default: os.Stdout)
	JSON    bool      // Use JSON format instead of text
}

// DefaultLogConfig returns the default logging configuration
// By default, logging is enabled with INFO level
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Enabled: true,
		Level:   LevelInfo,
		Writer:  os.Stdout,
		JSON:    false,
	}
}

// DisabledLogConfig returns a configuration with logging disabled
func DisabledLogConfig() *LogConfig {
	return &LogConfig{
		Enabled: false,
		Level:   LevelInfo,
		Writer:  io.Discard,
		JSON:    false,
	}
}

// Logger is the SDK logger interface
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	IsEnabled() bool
}

// sdkLogger is the implementation of the Logger interface using slog
type sdkLogger struct {
	slogLogger *slog.Logger
	enabled    bool
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config *LogConfig) Logger {
	if config == nil {
		config = DefaultLogConfig()
	}

	if !config.Enabled {
		config.Writer = io.Discard
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: convertLogLevel(config.Level),
	}

	if config.JSON {
		handler = slog.NewJSONHandler(config.Writer, opts)
	} else {
		handler = slog.NewTextHandler(config.Writer, opts)
	}

	return &sdkLogger{
		slogLogger: slog.New(handler),
		enabled:    config.Enabled,
	}
}

// convertLogLevel converts SDK LogLevel to slog.Level
func convertLogLevel(level LogLevel) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *sdkLogger) Debug(msg string, args ...any) {
	if l.enabled {
		l.slogLogger.Debug(msg, args...)
	}
}

func (l *sdkLogger) Info(msg string, args ...any) {
	if l.enabled {
		l.slogLogger.Info(msg, args...)
	}
}

func (l *sdkLogger) Warn(msg string, args ...any) {
	if l.enabled {
		l.slogLogger.Warn(msg, args...)
	}
}

func (l *sdkLogger) Error(msg string, args ...any) {
	if l.enabled {
		l.slogLogger.Error(msg, args...)
	}
}

func (l *sdkLogger) With(args ...any) Logger {
	return &sdkLogger{
		slogLogger: l.slogLogger.With(args...),
		enabled:    l.enabled,
	}
}

func (l *sdkLogger) IsEnabled() bool {
	return l.enabled
}

// Global default logger (can be replaced by users)
var defaultLogger Logger = NewLogger(DefaultLogConfig())

// SetDefaultLogger sets the global default logger for the SDK
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// GetDefaultLogger returns the current default logger
func GetDefaultLogger() Logger {
	return defaultLogger
}

// Helper functions for quick logging using the default logger
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}
