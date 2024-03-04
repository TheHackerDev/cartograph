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

// DataFilter is a filter object for data API requests.
type DataFilter struct {
	Accept TargetFilterSimple `json:"accept"`
	Ignore TargetFilterSimple `json:"ignore"`
	Return TargetFilterSimple `json:"return"`
}

// ReturnFilter determines what data to return, when used with the data API.
type ReturnFilter struct {
	URLSchemes bool `json:"url_schemes"`
	Hosts      bool `json:"hosts"`
	Paths      bool `json:"paths"`

	RequestTypes  bool `json:"request_types"`
	ResponseCodes bool `json:"response_codes"`

	Parameters      bool `json:"parameters"`
	RequestHeaders  bool `json:"request_headers"`
	ResponseHeaders bool `json:"response_headers"`

	DateFound bool `json:"date_found"`
	LastSeen  bool `json:"last_seen"`
}
