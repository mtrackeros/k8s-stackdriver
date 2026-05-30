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

package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestAltTokenSourceFetchesToken(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		expiry := time.Now().Add(time.Hour).Format(time.RFC3339Nano)
		fmt.Fprintf(w, `{"accessToken":"alternate-token","expireTime":%q}`, expiry)
	}))
	defer tokenServer.Close()

	source := &AltTokenSource{
		oauthClient: newAltTokenHTTPClient(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "source-token"})),
		tokenURL:    tokenServer.URL,
		tokenBody:   "{}",
	}

	token, err := source.token()
	if err != nil {
		t.Fatalf("source.token() returned error: %v", err)
	}
	if token.AccessToken != "alternate-token" {
		t.Fatalf("source.token() returned access token %q, want %q", token.AccessToken, "alternate-token")
	}
}

func TestAltTokenSourceDoesNotFollowRedirects(t *testing.T) {
	redirected := make(chan struct{}, 1)
	destination := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		redirected <- struct{}{}
	}))
	defer destination.Close()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer tokenServer.Close()

	source := &AltTokenSource{
		oauthClient: newAltTokenHTTPClient(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "source-token"})),
		tokenURL:    tokenServer.URL,
		tokenBody:   "{}",
	}

	if _, err := source.token(); err == nil {
		t.Fatal("source.token() returned nil error for non-successful response")
	}
	select {
	case <-redirected:
		t.Fatal("unexpected redirected request")
	default:
	}
}
