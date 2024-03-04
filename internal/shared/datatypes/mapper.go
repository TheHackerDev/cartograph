/*******************************************************************************
 * Copyright 2018-2024 Aaron Hnatiw
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 ******************************************************************************/

package datatypes

import (
	"net/url"
	"time"
)

// ReferrerData is a struct that holds the referer and destination URLs from a single HTTP request.
type ReferrerData struct {
	Referer     url.URL
	Destination url.URL
	Timestamp   time.Time
}

// MapperBrowserData holds mapper data sent from our browser scripts.
type MapperBrowserData struct {
	Source       string   `json:"source"`
	Destinations []string `json:"destinations"`
}
