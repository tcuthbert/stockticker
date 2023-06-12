package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	srv "github.com/tcuthbert/stockticker/webserver"
)

var (
	symbol string
	nDays  int

	apiBaseURL = "https://www.alphavantage.co"
	apiFunc    = "TIME_SERIES_DAILY_ADJUSTED"
	apiKeyFile = ""
	listenAddr = ":5000"
)

func setupDefaults() {
	var ok bool

	if symbol, ok = os.LookupEnv("SYMBOL"); !ok {
		symbol = "MSFT"
	}

	var err error

	nDays, err = strconv.Atoi(os.Getenv("NDAYS"))
	if err != nil {
		nDays = 10
	}
}

func makeAPIURL(apiKey string) (*url.URL, error) {
	params := url.Values{
		"symbol":   []string{symbol},
		"function": []string{apiFunc},
		"apikey":   []string{apiKey},
	}

	apiURL, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, err
	}

	apiURL = apiURL.JoinPath("query")
	apiURL.RawQuery = params.Encode()

	return apiURL, nil
}

func main() {
	setupDefaults()
	flag.StringVar(&listenAddr, "listen-addr", listenAddr, "server listen address")
	flag.StringVar(&apiBaseURL, "api-url", apiBaseURL, "url of the stock ticker API")
	flag.StringVar(&apiKeyFile, "api-keyfile", "", "file containing key data for the API")
	flag.IntVar(&nDays, "num-days", nDays, "last <num-days> of stock data")
	flag.StringVar(&symbol, "symbol", symbol, "stock symbol to lookup")
	flag.Parse()

	var (
		apiKey    string
		gotAPIKey bool
	)

	if apiKey, gotAPIKey = os.LookupEnv("APIKEY"); !gotAPIKey {
		if k, _ := os.ReadFile(apiKeyFile); cap(k) != 0 {
			apiKey = string(bytes.TrimSpace(k))
		} else {
			fmt.Fprintf(os.Stderr, "Refusing to start without api key data\n")
			os.Exit(1)
		}
	}

	apiURL, err := makeAPIURL(apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making API URL %s\n", err)
		os.Exit(1)
	}

	if err := srv.Start(&listenAddr, apiURL, nDays); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start server: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
