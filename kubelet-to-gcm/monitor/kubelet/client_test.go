/*
Copyright 2017 Google Inc.

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

package kubelet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-stackdriver/kubelet-to-gcm/monitor/util"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name        string
		host        string
		port        uint
		useAuthPort bool
		expectedURL string
	}{
		{
			name:        "IPv4 HTTP",
			host:        "127.0.0.1",
			port:        10255,
			useAuthPort: false,
			expectedURL: "http://127.0.0.1:10255/stats/summary",
		},
		{
			name:        "IPv4 HTTPS",
			host:        "127.0.0.1",
			port:        10250,
			useAuthPort: true,
			expectedURL: "https://127.0.0.1:10250/stats/summary",
		},
		{
			name:        "IPv6 HTTP",
			host:        "2001:db8::1",
			port:        10255,
			useAuthPort: false,
			expectedURL: "http://[2001:db8::1]:10255/stats/summary",
		},
		{
			name:        "IPv6 HTTPS",
			host:        "2001:db8::1",
			port:        10250,
			useAuthPort: true,
			expectedURL: "https://[2001:db8::1]:10250/stats/summary",
		},
		{
			name:        "Bracketed IPv6 HTTP",
			host:        "[2001:db8::1]",
			port:        10255,
			useAuthPort: false,
			expectedURL: "http://[2001:db8::1]:10255/stats/summary",
		},
		{
			name:        "Bracketed IPv6 HTTPS",
			host:        "[2001:db8::1]",
			port:        10250,
			useAuthPort: true,
			expectedURL: "https://[2001:db8::1]:10250/stats/summary",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient(tc.host, tc.port, http.DefaultClient, tc.useAuthPort)
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			if c.summaryURL.String() != tc.expectedURL {
				t.Errorf("Expected URL %q, got %q", tc.expectedURL, c.summaryURL.String())
			}
		})
	}
}

func TestDoRequestAndUnmarshal_SizeLimit(t *testing.T) {
	largeDataSize := maxResponseBodySize + 1024
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, largeDataSize)
		w.Write(data)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	k := &Client{
		client:     ts.Client(),
		summaryURL: u,
	}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	var value stats.Summary
	err := k.doRequestAndUnmarshal(ts.Client(), req, &value)

	if err == nil {
		t.Fatal("Expected error due to size limit, but got none")
	}

	if !strings.Contains(err.Error(), util.ErrBodyTooLarge.Error()) {
		t.Errorf("Expected error containing %q, got %v", util.ErrBodyTooLarge, err)
	}
}

func TestDoRequestAndUnmarshal_Success(t *testing.T) {
	summary := &stats.Summary{
		Node: stats.NodeStats{
			NodeName: "test-node",
		},
	}
	data, _ := json.Marshal(summary)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	k := &Client{
		client:     ts.Client(),
		summaryURL: u,
	}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	var result stats.Summary
	err := k.doRequestAndUnmarshal(ts.Client(), req, &result)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Node.NodeName != "test-node" {
		t.Errorf("Expected test-node, got %s", result.Node.NodeName)
	}
}
