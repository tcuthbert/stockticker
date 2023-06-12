package webserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/tcuthbert/stockticker/apiclient"
)

var ProcessedAPIResponse = `{"Meta Data":{"1. Information":"Daily Prices (open, high, low, close) and Volumes","2. Symbol":"IBM","3. Last Refreshed":"2023-06-08","4. Output Size":"Compact","5. Time Zone":"US/Eastern"},"Time Series (Daily)":{"2023-06-01":{"1. open":"128.44","2. high":"130.145","3. low":"127.78","4. close":"129.82","6. volume":"0"},"2023-06-02":{"1. open":"130.38","2. high":"133.12","3. low":"130.15","4. close":"132.42","6. volume":"0"},"2023-06-05":{"1. open":"133.12","2. high":"133.58","3. low":"132.27","4. close":"132.64","6. volume":"0"},"2023-06-06":{"1. open":"132.43","2. high":"132.94","3. low":"131.88","4. close":"132.69","6. volume":"0"},"2023-06-07":{"1. open":"132.5","2. high":"134.44","3. low":"132.19","4. close":"134.38","6. volume":"0"},"2023-06-08":{"1. open":"134.69","2. high":"135.98","3. low":"134.01","4. close":"134.41","6. volume":"0"},"2023-06-09":{"1. open":"134.6899","2. high":"135.9801","3. low":"134.0101","4. close":"134.4101","6. volume":"0"}},"ClosingAverage":"132.97"}`

func loadTestData() []byte {
	testFile, err := os.Open("testdata/api_response.json")
	if err != nil {
		return nil
	}
	defer testFile.Close()

	out, err := io.ReadAll(testFile)
	if err != nil {
		return nil
	}

	return out
}

func TestStart(t *testing.T) {
	listenAddr := "localhost:8080"
	apiURL, _ := url.Parse("https://api.example.com")
	numDays := 7

	err := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		err <- Start(&listenAddr, apiURL, numDays)
	}()

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("Unexpected error starting server: %v", ctx.Err())
		}
	case err := <-err:
		t.Errorf("Error starting server: %v", err)
	}
}

func TestHandlers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if _, err := rw.Write(loadTestData()); err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}))

	defer server.Close()

	fakeAPIClient := apiclient.NewAPIClient(server.Client(), server.URL)
	lh := log.Default()

	cases := []struct {
		w                    *httptest.ResponseRecorder
		r                    *http.Request
		th                   http.HandlerFunc
		expectedResponseCode int
		expectedResponseBody []byte
	}{
		{
			w:                    httptest.NewRecorder(),
			r:                    httptest.NewRequest("GET", server.URL, nil),
			th:                   reqHandler(lh, fakeAPIClient, 7, apiRequestHandler),
			expectedResponseCode: 200,
			expectedResponseBody: []byte(ProcessedAPIResponse),
		},
	}

	for _, c := range cases {
		c.th(c.w, c.r)

		if c.expectedResponseCode != c.w.Code {
			t.Errorf("Status Code didn't match:\n\t%q\n\t%q", c.expectedResponseCode, c.w.Code)
		}

		if !bytes.Equal(c.expectedResponseBody, c.w.Body.Bytes()) {
			t.Errorf("Body didn't match:\n\t%q\n\t%q", c.expectedResponseBody, c.w.Body.String())
		}
	}
}
