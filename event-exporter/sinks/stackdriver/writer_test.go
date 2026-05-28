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

package stackdriver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sd "google.golang.org/api/logging/v2"
)

func TestWriterEnablesPartialSuccess(t *testing.T) {
	type writeRequest struct {
		method string
		path   string
		body   *sd.WriteLogEntriesRequest
		err    error
	}

	requests := make(chan writeRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req sd.WriteLogEntriesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			requests <- writeRequest{
				method: r.Method,
				path:   r.URL.Path,
				err:    fmt.Errorf("decode request body: %w", err),
			}
		} else {
			requests <- writeRequest{
				method: r.Method,
				path:   r.URL.Path,
				body:   &req,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	service, err := sd.New(server.Client())
	if err != nil {
		t.Fatalf("New logging service: %v", err)
	}
	service.BasePath = server.URL + "/"

	writer := sdWriterImpl{service: service}
	writer.Write(
		[]*sd.LogEntry{{TextPayload: "entry"}},
		"projects/test-project/logs/events",
		&sd.MonitoredResource{Type: k8sCluster},
	)

	select {
	case req := <-requests:
		if req.err != nil {
			t.Fatal(req.err)
		}
		if req.path != "/v2/entries:write" {
			t.Fatalf("request path = %q, want %q", req.path, "/v2/entries:write")
		}
		if req.method != http.MethodPost {
			t.Fatalf("request method = %q, want %q", req.method, http.MethodPost)
		}
		if !req.body.PartialSuccess {
			t.Fatal("PartialSuccess = false, want true")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for write request")
	}
}
