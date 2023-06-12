package apiclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TODO: documentation
func TestAPIClient(t *testing.T) {
	testdata, err := os.ReadFile("testdata/api_response.json")
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if _, err := rw.Write(testdata); err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})

	server := httptest.NewServer(testHandler)
	defer server.Close()

	// TODO: rationale
	// TODO: do a call recording test
	// TODO: test to make sure api called with correct url

	apiClient := NewAPIClient(server.Client(), server.URL)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)

	resp, err := apiClient.DoAPIRequest(req)
	if err != nil {
		t.Errorf("Unexpected error: got %v, want %v\n", err, nil)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != string(testdata) {
		t.Errorf("Unexpected body: want %s\n", "OK")
	}
}
