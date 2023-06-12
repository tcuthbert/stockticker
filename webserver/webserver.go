package webserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/tcuthbert/stockticker/apiclient"
	"github.com/tcuthbert/stockticker/apiresponse"
)

var (
	MaxAPIResponseTimeout = 60 * time.Second
	MaxReadTimeout        = 15 * time.Second
	MaxWriteTimeout       = 30 * time.Second
	MaxIdleTimeout        = 120 * time.Second
)

type apiRequestHandlerFunc func(*http.Request, http.ResponseWriter, *apiclient.APIClient, chan error)

type apiResponseHandlerFunc func(*http.Response, error) error

type nDaysKey string

func Start(listenAddr *string, apiURL *url.URL, numDays int) error {
	logger := log.New(os.Stdout, "webserver: ", log.LstdFlags)

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	server := newWebserver(listenAddr, apiURL, numDays, logger)
	go gracefullShutdown(server, logger, quit, done)

	logger.Println(fmt.Sprintf("Server is ready to handle requests at: %s", *listenAddr))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("could not listen on %s: %w", *listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")

	return nil
}

func gracefullShutdown(server *http.Server, logger *log.Logger, quit <-chan os.Signal, done chan<- bool) {
	<-quit
	logger.Println("Server is shutting down...")

	shutDownTime := 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), shutDownTime)
	defer cancel()

	server.SetKeepAlivesEnabled(false)

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

	close(done)
}

func apiRoundTrip(apiClient *apiclient.APIClient, req *http.Request, h apiResponseHandlerFunc) error {
	resp, err := apiClient.DoAPIRequest(req)
	if err != nil {
		return fmt.Errorf("api client error: %w", err)
	}
	defer resp.Body.Close()

	return h(resp, err)
}

func apiRequestHandler(req *http.Request, rw http.ResponseWriter, apiClient *apiclient.APIClient, resultCh chan error) {
	err := apiRoundTrip(apiClient, req, func(resp *http.Response, err error) error {
		nDays, ok := req.Context().Value(nDaysKey("nDays")).(int)
		if !ok {
			return errors.New("request context missing nDays filter")
		}

		var apiResponse apiresponse.APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return fmt.Errorf("error decoding api response: %w", err)
		}

		apiResponse, err = apiResponse.FilteredByDays(nDays)
		if err != nil {
			return fmt.Errorf("error calculating closing average: %w", err)
		}

		out, err := apiResponse.Bytes()
		if err != nil {
			return err
		}

		if _, err := rw.Write(out); err != nil {
			http.Error(rw, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		}

		return nil
	})
	resultCh <- err
	close(resultCh) // TODO: is this needed given resultCh is buffered?
}

func reqHandler(lh *log.Logger, apiClient *apiclient.APIClient, nDays int, apiHandler apiRequestHandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := context.WithValue(r.Context(), nDaysKey("nDays"), nDays)
		ctx, cancel := context.WithTimeout(ctx, MaxAPIResponseTimeout) // TODO: mdn timeouts
		defer cancel()

		resultCh := make(chan error, 1)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiClient.APIURL, nil)
		if err != nil {
			resultCh <- err
			return
		}

		go apiHandler(req, rw, apiClient, resultCh)

		// TODO: logging format for time out/failed api requests
		select {
		case <-ctx.Done():
			lh.Printf("ERROR: request=%s response time=%s: %v\n", req.URL, time.Since(start), ctx.Err())
			http.Error(rw, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
		case err := <-resultCh:
			if err != nil {
				lh.Printf("ERROR: response time=%s: %v\n", time.Since(start), err)
				http.Error(rw, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			} else {
				lh.Printf("INFO: response time=%s\n", time.Since(start))
			}
		}
	}
}

func newWebserver(listenAddr *string, apiURL *url.URL, numDays int, logger *log.Logger) *http.Server {
	apiClient := apiclient.NewAPIClient(http.DefaultClient, apiURL.String())
	apiHandler := http.TimeoutHandler(reqHandler(logger, apiClient, numDays, apiRequestHandler), MaxAPIResponseTimeout, http.StatusText(http.StatusRequestTimeout))

	router := http.NewServeMux()

	router.Handle("/", apiHandler)

	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// TODO: use mdn recommended timeout values
	return &http.Server{
		Addr:         *listenAddr,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  MaxReadTimeout,
		WriteTimeout: MaxWriteTimeout,
		IdleTimeout:  MaxIdleTimeout,
	}
}
