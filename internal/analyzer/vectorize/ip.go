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

package vectorize

import (
	"errors"
	"net"
)

// ip.go contains the code to vectorize IP addresses.

// Ip converts the given IP address to a vector for machine learning analysis.
func Ip(ip net.IP) ([]float32, error) {
	// Convert IPv4 address to IPv6, if necessary
	ip.To16()
	if ip == nil {
		return nil, errors.New("invalid IP address")
	}

	// Create an integer slice with each decimal value in the IPv6 address
	decimalSlice := make([]float32, len(ip))
	for i, v := range ip {
		decimalSlice[i] = float32(int(v))
	}

	return decimalSlice, nil
}
