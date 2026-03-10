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

package constants

import (
	"runtime/debug"
	"strings"
)

// Version can be set at build time using ldflags:
// go build -ldflags "-X github.com/PhonePe/phonepe-pg-sdk-go/common/constants.Version=v1.0.0"
var Version = "dev"

// SubscriptionVersion can be set at build time using ldflags (defaults to same as Version):
// go build -ldflags "-X github.com/PhonePe/phonepe-pg-sdk-go/common/constants.SubscriptionVersion=v2.1.7"
var SubscriptionVersion = ""

// getSDKVersion returns the current SDK version
// It first checks if Version was set at build time, then falls back to build info
func getSDKVersion() string {
	// If version was set via ldflags
	if Version != "" && Version != "dev" {
		return Version
	}

	// Try to get version from build info (works with go modules)
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return strings.TrimPrefix(info.Main.Version, "v")
		}
	}

	// Fallback to default
	return "1.0.0"
}

// getSubscriptionAPIVersion returns the subscription API version
// Uses SubscriptionVersion if set, otherwise falls back to SDK version
func getSubscriptionAPIVersion() string {
	// If subscription version was explicitly set via ldflags
	if SubscriptionVersion != "" && SubscriptionVersion != "dev" {
		return SubscriptionVersion
	}

	// Otherwise use the same version as SDK
	return getSDKVersion()
}

// SDKVersion uses the version detection logic
var SDKVersion = getSDKVersion()

// SubscriptionAPIVersion uses the subscription version detection logic
// By default, it will be the same as SDKVersion unless explicitly set at build time
var SubscriptionAPIVersion = getSubscriptionAPIVersion()

const (
	APIVersion            = "V2"
	Integration           = "API"
	SDKType               = "BACKEND_GO_SDK" // Changed from JAVA to GO
	Source                = "x-source"
	SourceVersion         = "x-source-version"
	SourcePlatform        = "x-source-platform"
	SourcePlatformVersion = "x-source-platform-version"
	OAuthAuthorization    = "Authorization"
	ContentType           = "Content-Type"
	Accept                = "Accept"
)
