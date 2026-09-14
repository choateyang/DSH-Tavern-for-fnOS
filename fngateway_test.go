package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestGatewayBackendPath(t *testing.T) {
	tests := map[string]string{
		"/app/dsh-tavern/fngateway/":               "/",
		"/app/dsh-tavern/fngateway/plugins/":       "/plugins/",
		"/fngateway/plugins/":                      "/plugins/",
		"/app/dsh-tavern/fngateway/api/remote.mux": "/api/remote.mux",
		"/app/dsh-tavern/fngateway":                "/",
		"/other/path":                              "/other/path",
	}
	for input, want := range tests {
		if got := gatewayBackendPath(input); got != want {
			t.Fatalf("gatewayBackendPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRestoreDshComboQuery(t *testing.T) {
	combo := "@deepseek-ai/dsh-client-modules/client.js&rev=abc123"
	if got := restoreDshComboQuery("/plugins/", combo); got != "?"+combo {
		t.Fatalf("restored combo query = %q, want %q", got, "?"+combo)
	}
	if got := restoreDshComboQuery("/plugins/", "?"+combo); got != "?"+combo {
		t.Fatalf("existing combo query changed to %q", got)
	}
	if got := restoreDshComboQuery("/plugins/item/client.js", "rev=abc123"); got != "rev=abc123" {
		t.Fatalf("normal plugin query changed to %q", got)
	}
}

func TestGatewayProxyRewritePreservesComboRequest(t *testing.T) {
	const combo = "@deepseek-ai/dsh-client-modules/client.js&rev=abc123"
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.RequestURI(), "/plugins/??"+combo; got != want {
			t.Errorf("backend received %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	target, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	proxy := &httputil.ReverseProxy{Rewrite: func(pr *httputil.ProxyRequest) {
		pr.SetURL(target)
		rewriteFnGatewayURL(pr.In.URL, pr.Out.URL)
	}}
	req := httptest.NewRequest(http.MethodGet, fnGatewayPrefix+"/plugins/??"+combo, nil)
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("proxy status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
