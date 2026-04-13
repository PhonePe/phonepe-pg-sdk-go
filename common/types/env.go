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

package types

import (
	"github.com/PhonePe/phonepe-pg-sdk-go/common/constants"
)

type Env struct {
	PgHostURL     string
	PciPgHostURL  string
	OAuthHostURL  string
	EventsHostURL string
}

var (
	Sandbox = Env{
		PgHostURL:     constants.SandboxPgHostURL,
		PciPgHostURL:  constants.SandboxPgHostURL,
		OAuthHostURL:  constants.SandboxOAuthHostURL,
		EventsHostURL: constants.SandboxEventsHostURL,
	}
	Production = Env{
		PgHostURL:     constants.ProductionPgHostURL,
		PciPgHostURL:  constants.ProductionPciPgHostURL,
		OAuthHostURL:  constants.ProductionOAuthHostURL,
		EventsHostURL: constants.ProductionEventsHostURL,
	}
	Test = Env{
		PgHostURL:     constants.TestingURL,
		PciPgHostURL:  constants.TestingURL,
		OAuthHostURL:  constants.TestingURL,
		EventsHostURL: constants.TestingURL,
	}
)

// NewEnv creates a custom Env. If pciPgHostURL is empty, it defaults to pgHostURL
// (sandbox behaviour: no separate PCI zone).
func NewEnv(pgHostURL, pciPgHostURL, oAuthHostURL, eventsHostURL string) Env {
	if pciPgHostURL == "" {
		pciPgHostURL = pgHostURL
	}
	return Env{
		PgHostURL:     pgHostURL,
		PciPgHostURL:  pciPgHostURL,
		OAuthHostURL:  oAuthHostURL,
		EventsHostURL: eventsHostURL,
	}
}
