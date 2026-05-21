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
	"bytes"
	"strings"
	"testing"
)

func TestReadWithLimit(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		limit       int64
		expected    string
		expectedErr error
	}{
		{
			name:        "under limit",
			input:       "short",
			limit:       10,
			expected:    "short",
			expectedErr: nil,
		},
		{
			name:        "at limit",
			input:       "exactlimit",
			limit:       10,
			expected:    "exactlimit",
			expectedErr: nil,
		},
		{
			name:        "over limit",
			input:       "this is way too long",
			limit:       10,
			expected:    "",
			expectedErr: ErrBodyTooLarge,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReadWithLimit(strings.NewReader(tc.input), tc.limit)
			if err != tc.expectedErr {
				t.Fatalf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if tc.expectedErr == nil && !bytes.Equal(got, []byte(tc.expected)) {
				t.Errorf("Expected data %q, got %q", tc.expected, string(got))
			}
		})
	}
}
