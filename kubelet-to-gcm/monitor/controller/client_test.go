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

package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-stackdriver/kubelet-to-gcm/monitor/util"
)

func TestDoRequestAndParse_SizeLimit(t *testing.T) {
	// Create a mock server that returns a response larger than maxResponseBodySize.
	largeDataSize := maxResponseBodySize + 1024
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write more than maxResponseBodySize bytes.
		data := make([]byte, largeDataSize)
		w.Write(data)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := &Client{
		client:     ts.Client(),
		metricsURL: u,
	}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	_, err := client.doRequestAndParse(req)

	if err == nil {
		t.Fatal("Expected error due to size limit, but got none")
	}

	if !strings.Contains(err.Error(), util.ErrBodyTooLarge.Error()) {
		t.Errorf("Expected error containing %q, got %v", util.ErrBodyTooLarge, err)
	}
}

func TestDoRequestAndParse_Success(t *testing.T) {
	metricsData := `
# HELP node_collector_evictions_number Number of evictions
# TYPE node_collector_evictions_number counter
node_collector_evictions_number 10
# HELP process_start_time_seconds Start time
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1234567890
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(metricsData))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := &Client{
		client:     ts.Client(),
		metricsURL: u,
	}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	metrics, err := client.doRequestAndParse(req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if metrics.NodeEvictions != 10 {
		t.Errorf("Expected 10 evictions, got %d", metrics.NodeEvictions)
	}
	if metrics.CreateTime != 1234567890 {
		t.Errorf("Expected 1234567890 create time, got %d", metrics.CreateTime)
	}
}
