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

package http

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
)

// DecodeGzip returns a decoded version of the given gzip-encoded contents.
func DecodeGzip(contents []byte) ([]byte, error) {
	br := bytes.NewReader(contents)
	decompressor, newReaderErr := gzip.NewReader(br)
	if newReaderErr != nil {
		return nil, fmt.Errorf("unable to create new gzip reader to read contents: %w", newReaderErr)
	}
	decoded, decodeErr := io.ReadAll(decompressor)
	if decodeErr != nil {
		return nil, fmt.Errorf("unable to decode gzip-encoded contents: %w", decodeErr)
	}
	return decoded, nil
}

// DecodeBrotli returns a decoded version of the given brotli-encoded contents
func DecodeBrotli(contents []byte) ([]byte, error) {
	br := bytes.NewReader(contents)
	decompressor := brotli.NewReader(br)
	decoded, decodeErr := io.ReadAll(decompressor)
	if decodeErr != nil {
		return nil, fmt.Errorf("unable to decode brotli-encoded contents: %w", decodeErr)
	}
	return decoded, nil

}

// DecodeDeflate returns a decoded version of the given deflate-encoded contents
func DecodeDeflate(contents []byte) ([]byte, error) {
	br := bytes.NewReader(contents)
	decompressor := flate.NewReader(br)
	decoded, decodeErr := io.ReadAll(decompressor)
	if decodeErr != nil {
		return nil, fmt.Errorf("unable to decode deflate-encoded contents: %w", decodeErr)
	}
	return decoded, nil

}
