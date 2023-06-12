package apiresponse

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"time"
)

const DATEFORMAT = time.DateOnly

func roundFloat(f float64) float64 {
	return math.Round(f*100.0) / 100.0
}

type MetaData struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

type DailyTimeSeries struct {
	Open   float64 `json:"1. open,string"`
	High   float64 `json:"2. high,string"`
	Low    float64 `json:"3. low,string"`
	Close  float64 `json:"4. close,string"`
	Volume int64   `json:"6. volume,string"`
}

type APIResponse struct {
	MetaData       MetaData                   `json:"Meta Data"`
	TimeSeriesData map[string]DailyTimeSeries `json:"Time Series (Daily)"`
	ClosingAverage float64                    `json:",string"`
}

func NewAPIResponse(r io.Reader) (APIResponse, error) {
	var apiResponse APIResponse

	if err := json.NewDecoder(r).Decode(&apiResponse); err != nil {
		return APIResponse{}, fmt.Errorf("error decoding api response: %w", err)
	}

	return apiResponse.WithClosingAverage(), nil
}

func (apr APIResponse) Bytes() ([]byte, error) {
	out, err := json.Marshal(apr)
	if err != nil {
		err = fmt.Errorf("error encoding API response: %w", err)
	}

	return out, err
}

func (apr APIResponse) WithClosingAverage() APIResponse {
	count := float64(len(apr.TimeSeriesData))
	total := float64(0)

	for _, v := range apr.TimeSeriesData {
		total += v.Close
	}

	avg := total / count
	apr.ClosingAverage = roundFloat(avg)

	return apr
}

func (apr APIResponse) FilteredByDays(nDays int) (APIResponse, error) {
	keys := make([]string, 0, len(apr.TimeSeriesData))

	for k := range apr.TimeSeriesData {
		if _, err := time.Parse(DATEFORMAT, k); err != nil {
			return APIResponse{}, fmt.Errorf("error parsing %s as datetime %s: %w", k, DATEFORMAT, err)
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)

	newTSData := make(map[string]DailyTimeSeries)

	counter := 0
	for idx := len(keys) - 1; idx >= 0; idx-- {
		if counter >= nDays {
			break
		}
		counter++

		k := keys[idx]
		v := apr.TimeSeriesData[k]
		newTSData[k] = v
	}

	apr.TimeSeriesData = newTSData

	return apr.WithClosingAverage(), nil
}
