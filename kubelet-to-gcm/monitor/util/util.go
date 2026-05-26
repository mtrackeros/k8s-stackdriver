/*
Copyright 2026 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"errors"
	"io"
	"strings"
)

var (
	// ErrBodyTooLarge is returned when the response body exceeds the configured limit.
	ErrBodyTooLarge = errors.New("response body too large")
)

// ReadWithLimit reads from r until EOF or the limit is reached.
// If the limit is exceeded, it returns ErrBodyTooLarge.
func ReadWithLimit(r io.Reader, limit int64) ([]byte, error) {
	// Read up to limit + 1 bytes to detect overflow.
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, ErrBodyTooLarge
	}
	return data, nil
}

// NormalizeHost handles host strings that may already be bracketed (like IPv6).
// net.JoinHostPort expects a raw IP or hostname; if it receives a bracketed IPv6,
// it will double-bracket it.
func NormalizeHost(host string) string {
	host = strings.TrimSpace(host)
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host[1 : len(host)-1]
	}
	return host
}
