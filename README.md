# PhonePe B2B Payment Gateway SDK for Go

![Go](https://img.shields.io/badge/Go-1.21%2B-blue)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

A Go library for seamless integration with PhonePe Payment Gateway APIs.

## Table of Contents
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Initialization](#initialization)
  - [Standard Checkout Flow](#standard-checkout-flow)
  - [Checking Order Status](#checking-order-status)
  - [Handling Callbacks](#handling-callbacks)
  - [SDK Order Integration](#sdk-order-integration)
- [Advanced Configuration](#advanced-configuration)
  - [Context and Timeouts](#context-and-timeouts)
  - [Logging Configuration](#logging-configuration)
  - [Retry Configuration](#retry-configuration)
- [Contributing](#contributing)
- [Documentation](#documentation)
- [License](#license)

## Requirements

- Go 1.21 or later

## Installation

Add the dependency to your project:

```bash
go get github.com/PhonePe/phonepe-pg-sdk-go
```

Or add it to your `go.mod` file:

```go
require github.com/PhonePe/phonepe-pg-sdk-go v1.0.0
```

## Quick Start

### Initialization

Before using the SDK, you need to acquire your credentials from the [PhonePe Merchant Portal](https://developer.phonepe.com/v1/docs/merchant-onboarding).

You need three key pieces of information:
1. `clientId` - Your merchant identifier
2. `clientSecret` - Your authentication secret
3. `clientVersion` - API version to use

```go
import (
    "context"
    "time"
    "github.com/PhonePe/phonepe-pg-sdk-go/common/types"
    "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/standardcheckout"
)

clientId := "<your-client-id>"
clientSecret := "<your-client-secret>"
clientVersion := 1           // Your client version here
env := types.Sandbox         // Use types.Production for live transactions
shouldPublishEvents := false // Set to true for production

standardCheckoutClient, err := standardcheckout.GetInstance(
    clientId,
    clientSecret,
    clientVersion,
    env,
    shouldPublishEvents,
)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

### Standard Checkout Flow

To initiate a payment, create a request using `NewStandardCheckoutPayRequest`:

```go
import (
    "context"
    "time"
    "github.com/google/uuid"
    "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
    "github.com/PhonePe/phonepe-pg-sdk-go/common/models"
)

// Generate a unique order ID
merchantOrderId := uuid.New().String()
amount := int64(10000) // Amount in lowest currency denomination (paise for INR)
redirectUrl := "https://www.yourwebsite.com/redirect"

payRequest := request.NewStandardCheckoutPayRequest(
    merchantOrderId,
    amount,
    redirectUrl,
    models.MetaInfo{},
    "Payment for " + merchantOrderId,
    1800, // Expire after 30 minutes
    nil,  // paymentModeConfig (optional)
    nil,  // disablePaymentRetry (optional)
)

// Create a context with timeout for the API call
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

payResponse, err := standardCheckoutClient.Pay(ctx, payRequest)
if err != nil {
    log.Fatalf("Payment initiation failed: %v", err)
}

checkoutPageUrl := payResponse.RedirectURL

// Redirect the user to checkoutPageUrl to complete the payment
```

### Checking Order Status

To check the status of an order:

```go
merchantOrderId := "<your-merchant-order-id>" // Order ID created during payment initialization

// Create a context with timeout for the API call
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

orderStatusResponse, err := standardCheckoutClient.GetOrderStatus(ctx, merchantOrderId)
if err != nil {
    log.Fatalf("Failed to get order status: %v", err)
}

state := orderStatusResponse.Data.State

// Handle the state accordingly in your application
```

### Handling Callbacks

PhonePe sends callbacks to your configured endpoint. Validate these callbacks to ensure they're authentic:

```go
import (
    "github.com/PhonePe/phonepe-pg-sdk-go/common/models/response"
)

// Credentials for SHA256 authentication that you've configured in the PhonePe dashboard
username := "<your-username>"
password := "<your-password>"

// Data received in the callback (from your HTTP handler)
authorization := r.Header.Get("Authorization") // SHA256 hash
responseBody, _ := io.ReadAll(r.Body)          // JSON body as string

callbackResponse, err := standardCheckoutClient.ValidateCallback(
    username,
    password,
    authorization,
    string(responseBody),
)
if err != nil {
    // Handle invalid callback - potential security issue
    log.Printf("Invalid callback: %v", err)
    return
}

orderId := callbackResponse.Data.MerchantOrderID
state := callbackResponse.Data.State

// Process the order based on its state
```

Possible callback states include:
- `COMPLETED` - Payment completed successfully
- `FAILED` - Payment failed
- `PENDING` - Payment is still in progress

### SDK Order Integration

For mobile SDK integration, first create an order on your server:

```go
import (
    "context"
    "time"
    "github.com/google/uuid"
    "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
    "github.com/PhonePe/phonepe-pg-sdk-go/common/models"
)

merchantOrderId := uuid.New().String()
amount := int64(10000) // Amount in lowest denomination (paise for INR)
redirectUrl := "https://yourapp.com/callback"

orderRequest := request.NewStandardCheckoutCreateSdkOrderRequest(
    amount,
    merchantOrderId,
    models.MetaInfo{},
    "Payment for "+merchantOrderId,
    redirectUrl,
    1800, // Expire after 30 minutes
    nil,  // disablePaymentRetry (optional)
    nil,  // paymentModeConfig (optional)
)

// Create a context with timeout for the API call
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

orderResponse, err := standardCheckoutClient.CreateSdkOrder(ctx, orderRequest)
if err != nil {
    log.Fatalf("Failed to create SDK order: %v", err)
}

token := orderResponse.OrderID

// Pass this token to your mobile app to initiate payment through the PhonePe SDK
```

## Advanced Configuration

### Context and Timeouts

All SDK methods accept a `context.Context` as the first parameter, allowing you to:
- **Cancel requests** when users navigate away
- **Set timeouts** per request
- **Propagate trace IDs** for distributed tracing
- **Handle graceful shutdowns**

**Basic timeout example:**
```go
// Create a context with 10-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

response, err := standardCheckoutClient.Pay(ctx, payRequest)
if err != nil {
    if err == context.DeadlineExceeded {
        log.Println("Request timed out")
    }
    return err
}
```

**Cancellation example:**
```go
// Create a cancellable context
ctx, cancel := context.WithCancel(context.Background())

// Cancel if user navigates away
go func() {
    <-userClosedBrowser
    cancel()
}()

response, err := standardCheckoutClient.GetOrderStatus(ctx, orderId)
```

**Best practices:**
- Always use `defer cancel()` after creating a context
- Use `context.WithTimeout()` for all production API calls
- Typical timeout: 10-30 seconds for payment operations
- Check for `context.DeadlineExceeded` and `context.Canceled` errors

### Logging Configuration

The SDK supports comprehensive structured logging with configurable log levels. By default, logging is enabled at INFO level.

**Disable Logging:**
```go
import (
    "github.com/PhonePe/phonepe-pg-sdk-go/common"
    "github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
)

client, err := common.NewBaseClientWithOptions(
    clientId,
    clientSecret,
    clientVersion,
    env,
    shouldPublishEvents,
    nil, // retry config
    logger.DisabledLogConfig(),
)
```

**Debug Level Logging:**
```go
import "os"

debugConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelDebug,
    Writer:  os.Stdout,
    JSON:    false,
}

client, err := common.NewBaseClientWithOptions(
    clientId,
    clientSecret,
    clientVersion,
    env,
    shouldPublishEvents,
    nil,
    debugConfig,
)
```

**Production Logging (JSON format, ERROR level only):**
```go
logFile, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

prodConfig := &logger.LogConfig{
    Enabled: true,
    Level:   logger.LevelError,
    Writer:  logFile,
    JSON:    true,
}
```

**Log Levels:**
- `logger.LevelDebug` - Detailed diagnostic information
- `logger.LevelInfo` - General informational messages (default)
- `logger.LevelWarn` - Warning messages
- `logger.LevelError` - Error conditions

For comprehensive logging documentation, see [LOGGING_GUIDE.md](LOGGING_GUIDE.md).

### Retry Configuration

Configure automatic retries for transient failures:

```go
import (
    "github.com/PhonePe/phonepe-pg-sdk-go/common"
    commonHttp "github.com/PhonePe/phonepe-pg-sdk-go/common/http"
)

// Conservative retry (2 attempts, short delays)
retryConfig := commonHttp.ConservativeRetryConfig()

// Aggressive retry (5 attempts, exponential backoff)
retryConfig := commonHttp.AggressiveRetryConfig()

// Custom retry configuration
retryConfig := &commonHttp.RetryConfig{
    Enabled:        true,
    MaxAttempts:    3,
    InitialDelay:   100,  // milliseconds
    MaxDelay:       5000, // milliseconds
    Multiplier:     2.0,  // exponential backoff
    RetryableStatusCodes: []int{429, 500, 502, 503, 504},
}

client, err := common.NewBaseClientWithRetry(
    clientId,
    clientSecret,
    clientVersion,
    env,
    shouldPublishEvents,
    retryConfig,
)
```

**Combine Logging and Retry:**
```go
client, err := common.NewBaseClientWithOptions(
    clientId,
    clientSecret,
    clientVersion,
    env,
    shouldPublishEvents,
    retryConfig,     // retry configuration
    logConfig,       // logging configuration
)
```

## Documentation

For detailed API documentation, advanced features, and integration options:

- [Standard Checkout Documentation](https://developer.phonepe.com/v1/reference/standard-checkout-introduction)
- [Subscription Documentation](https://developer.phonepe.com/v1/reference/autopay-introduction)
- [PhonePe Developer Portal](https://developer.phonepe.com/)
- [Logging Guide](LOGGING_GUIDE.md) - Comprehensive guide to SDK logging

## Contributing

Contributions to PG Go SDK are welcome! Here's how you can contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure your code follows the project's coding standards and includes appropriate tests.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

```
Copyright 2026 PhonePe Private Limited

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
