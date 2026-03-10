# Logging Guide

## Overview

The PhonePe Go SDK now includes comprehensive structured logging capabilities built on Go's standard `log/slog` package (Go 1.21+). This guide explains how to configure and use logging in your applications.

## Key Features

✅ **Structured Logging** - Use key-value pairs for better log analysis  
✅ **Log Levels** - DEBUG, INFO, WARN, ERROR  
✅ **Configurable** - Enable/disable logging and set minimum log level  
✅ **Multiple Formats** - Text and JSON output formats  
✅ **Zero Dependencies** - Uses Go's standard library `log/slog`  
✅ **Context Support** - Add contextual information to all logs  

## Quick Start

### Default Logging (INFO level)

```go
import (
    "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/standardcheckout"
    "github.com/PhonePe/phonepe-pg-sdk-go/common/types"
)

// Default: INFO level, text format, enabled
client, err := standardcheckout.GetInstance(
    "client-id",
    "client-secret",
    1,
    types.Sandbox,
    false,
)
// Logs: INFO, WARN, ERROR messages
```

### Disable Logging

```go
import (
    "github.com/PhonePe/phonepe-pg-sdk-go/common"
    "github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
)

client, err := common.NewBaseClientWithOptions(
    "client-id",
    "client-secret",
    1,
    types.Sandbox,
    false,
    nil, // no retry config
    logger.DisabledLogConfig(),
)
// No logs will be output
```

### Debug Level Logging

```go
import (
    "os"
    "github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
)

debugConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelDebug,
    Writer:  os.Stdout,
    JSON:    false,
}

client, err := common.NewBaseClientWithOptions(
    "client-id",
    "client-secret",
    1,
    types.Sandbox,
    false,
    nil,
    debugConfig,
)
// Logs: DEBUG, INFO, WARN, ERROR messages
```

### JSON Format Logging

```go
jsonConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelInfo,
    Writer:  os.Stdout,
    JSON:    true, // Enable JSON format
}

client, err := common.NewBaseClientWithOptions(
    "client-id",
    "client-secret",
    1,
    types.Sandbox,
    false,
    nil,
    jsonConfig,
)
// Output: {"time":"2026-02-10T10:00:00Z","level":"INFO","msg":"..."}
```

## Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| **DEBUG** | Detailed information for diagnosis | Development, troubleshooting |
| **INFO** | General informational messages | Normal operations, default level |
| **WARN** | Warning messages that need attention | Non-critical issues, fallbacks |
| **ERROR** | Error messages for failures | Failures, exceptions |

### Level Filtering

When you set a log level, only messages at that level or higher will be output:

```go
// Level: ERROR - only ERROR messages appear
// Level: WARN  - WARN and ERROR messages appear
// Level: INFO  - INFO, WARN, and ERROR messages appear (default)
// Level: DEBUG - All messages appear
```

## LogConfig Options

```go
type LogConfig struct {
    Enabled bool      // Enable or disable logging
    Level   LogLevel  // Minimum log level (DEBUG, INFO, WARN, ERROR)
    Writer  io.Writer // Output destination (os.Stdout, os.Stderr, file, etc.)
    JSON    bool      // Use JSON format instead of text
}
```

### Examples

**Production Configuration (ERROR only, JSON):**
```go
prodConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelError,
    Writer:  os.Stderr,
    JSON:    true,
}
```

**Development Configuration (DEBUG, text):**
```go
devConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelDebug,
    Writer:  os.Stdout,
    JSON:    false,
}
```

**Log to File:**
```go
logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}
defer logFile.Close()

fileConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelInfo,
    Writer:  logFile,
    JSON:    true,
}
```

## Structured Logging

Add contextual information using key-value pairs:

```go
// Simple message
client.Logger.Info("payment initiated")

// With structured fields
client.Logger.Info("payment initiated",
    "order_id", "ORDER123",
    "amount", 10000,
    "currency", "INR",
)

// Output (text format):
// time=2026-02-10T10:00:00.000+00:00 level=INFO msg="payment initiated" order_id=ORDER123 amount=10000 currency=INR

// Output (JSON format):
// {"time":"2026-02-10T10:00:00Z","level":"INFO","msg":"payment initiated","order_id":"ORDER123","amount":10000,"currency":"INR"}
```

## Contextual Logging

Create a logger with pre-populated context:

```go
// Add context that appears in all subsequent logs
contextLogger := client.Logger.With(
    "component", "payment",
    "merchant_id", "MERCHANT123",
)

contextLogger.Info("payment started", "amount", 10000)
// Output: ... component=payment merchant_id=MERCHANT123 msg="payment started" amount=10000

contextLogger.Error("payment failed", "error", "timeout")
// Output: ... component=payment merchant_id=MERCHANT123 msg="payment failed" error=timeout
```

## SDK Internal Logging

The SDK logs key events automatically:

### Token Service
- **DEBUG**: "returning cached token"
- **WARN**: "returning cached token, error occurred while fetching new token"
- **ERROR**: "no cached token available, error occurred while fetching new token"
- **INFO**: "force refreshing token"
- **ERROR**: "recovered from panic during force refresh"
- **ERROR**: "error refreshing token"

### Usage Example

```go
// Enable DEBUG logging to see token caching behavior
debugConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelDebug,
    Writer:  os.Stdout,
    JSON:    false,
}

client, _ := standardcheckout.GetInstanceWithOptions(
    "client-id",
    "client-secret",
    1,
    types.Sandbox,
    false,
    nil,
    debugConfig,
)

// First payment - token fetch
response1, _ := client.Pay(payRequest)
// Log: ... level=INFO msg="fetching OAuth token"

// Second payment - cached token
response2, _ := client.Pay(payRequest)
// Log: ... level=DEBUG msg="returning cached token"
```

## Best Practices

### 1. Use Appropriate Log Levels in Production

```go
// Production: ERROR level to reduce noise
prodConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelError,
    Writer:  os.Stderr,
    JSON:    true,
}
```

### 2. Enable DEBUG Only in Development

```go
var logConfig *logger.LogConfig
if os.Getenv("ENV") == "production" {
    logConfig = &logger.LogConfig{
        Enabled: true,
        Level:   logger.LevelError,
        Writer:  os.Stderr,
        JSON:    true,
    }
} else {
    logConfig = &logger.LogConfig{
        Enabled: true,
        Level:   logger.LevelDebug,
        Writer:  os.Stdout,
        JSON:    false,
    }
}
```

### 3. Use Structured Fields

```go
// ❌ Bad: String formatting loses structure
client.Logger.Error(fmt.Sprintf("Payment failed for order %s with error %v", orderID, err))

// ✅ Good: Structured fields for better analysis
client.Logger.Error("payment failed",
    "order_id", orderID,
    "error", err,
    "retry_count", retryCount,
)
```

### 4. Add Context for Related Operations

```go
// Create contextual logger for a transaction
txLogger := client.Logger.With(
    "transaction_id", txID,
    "user_id", userID,
)

txLogger.Info("transaction started")
txLogger.Info("validating payment details")
txLogger.Info("calling payment gateway")
txLogger.Info("transaction completed")
// All logs will include transaction_id and user_id
```

### 5. Disable Logging for Tests

```go
func TestPaymentFlow(t *testing.T) {
    client, _ := common.NewBaseClientWithOptions(
        "test-id",
        "test-secret",
        1,
        types.Sandbox,
        false,
        nil,
        logger.DisabledLogConfig(), // Disable logs in tests
    )
    // ... test code
}
```

## Migration from Old Logging

If you were using standard `log` package:

**Before:**
```go
import "log"

log.Println("Returning cached token")
log.Printf("Error: %v", err)
```

**After:**
```go
// SDK handles this internally now
client.Logger.Debug("returning cached token")
client.Logger.Error("error occurred", "error", err)
```

## Examples

See the complete examples in:
- `examples/logging_example.go`

Run the example:
```bash
go run examples/logging_example.go
```

## Troubleshooting

**Q: Logs are not appearing**  
A: Check that logging is enabled and the log level is appropriate for your messages.

**Q: Too many logs in production**  
A: Use `LevelError` or `LevelWarn` in production environments.

**Q: Need to log to multiple destinations**  
A: Use `io.MultiWriter` to write to multiple outputs:
```go
multiWriter := io.MultiWriter(os.Stdout, logFile)
config := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelInfo,
    Writer:  multiWriter,
    JSON:    true,
}
```

**Q: Want to integrate with existing logging framework**  
A: Implement the `logger.Logger` interface with your custom logger and pass it to the SDK.

## Summary

The PhonePe Go SDK provides flexible, production-ready logging:
- ✅ Enable/disable as needed
- ✅ Control verbosity with log levels
- ✅ Structured logging for better analysis
- ✅ JSON format for log aggregation tools
- ✅ Zero external dependencies

For more information, see the [API documentation](https://pkg.go.dev/github.com/PhonePe/phonepe-pg-sdk-go/common/logger).
