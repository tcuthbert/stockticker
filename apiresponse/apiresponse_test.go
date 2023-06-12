package apiresponse

import (
	"strings"
	"testing"
)

func jsonData() string {
	return `{
    "Meta Data": {
      "1. Information": "Sample Information",
      "2. Symbol": "Sample Symbol",
      "3. Last Refreshed": "2023-06-10",
      "4. Output Size": "Compact",
      "5. Time Zone": "UTC"
    },
    "Time Series (Daily)": {
      "2023-06-10": {
        "1. open": "100.50",
        "2. high": "105.20",
        "3. low": "99.80",
        "4. close": "102.30",
        "6. volume": "1000"
      },
      "2023-06-09": {
        "1. open": "98.50",
        "2. high": "101.20",
        "3. low": "97.80",
        "4. close": "99.30",
        "6. volume": "2000"
      },
      "2023-06-08": {
        "1. open": "98.50",
        "2. high": "101.20",
        "3. low": "97.80",
        "4. close": "80.30",
        "6. volume": "2000"
      },
      "2023-06-07": {
        "1. open": "98.50",
        "2. high": "101.20",
        "3. low": "97.80",
        "4. close": "103.20",
        "6. volume": "2000"
      }
    }
  }`
}

func TestAPIResponseWithClosingAverage(t *testing.T) {
	jsonStr := jsonData()

	reader := strings.NewReader(jsonStr)

	apiResponse, err := NewAPIResponse(reader)
	if err != nil {
		t.Errorf("NewAPIResponse returned an error: %v", err)
		return
	}

	// Assert that the calculated closing average is correct
	expectedAvg := 96.28
	if got := apiResponse.ClosingAverage; got != expectedAvg {
		t.Errorf("Expected ClosingAverage to be %.2f, got %.2f", expectedAvg, got)
	}
}

func TestNewAPIResponse(t *testing.T) {
	jsonStr := jsonData()

	reader := strings.NewReader(jsonStr)

	apiResponse, err := NewAPIResponse(reader)
	if err != nil {
		t.Errorf("NewAPIResponse returned an error: %v", err)
		return
	}

	// Assert that the API response has been parsed correctly
	expectedInfo := "Sample Information"
	if apiResponse.MetaData.Information != expectedInfo {
		t.Errorf("Expected MetaData.Information to be %q, got %q", expectedInfo, apiResponse.MetaData.Information)
	}
}

func TestAPIResponse(t *testing.T) {
	testCases := []struct {
		name     string
		apiResp  APIResponse
		nDays    int
		expected int
	}{
		{
			name: "Filter 3 days",
			apiResp: APIResponse{
				MetaData: MetaData{},
				TimeSeriesData: map[string]DailyTimeSeries{
					"2023-06-10": {},
					"2023-06-09": {},
					"2023-06-08": {},
					"2023-06-07": {},
				},
			},
			nDays:    3,
			expected: 3,
		},
		{
			name: "Filter 7 days",
			apiResp: APIResponse{
				MetaData: MetaData{},
				TimeSeriesData: map[string]DailyTimeSeries{
					"2023-06-10": {},
					"2023-06-09": {},
					"2023-06-08": {},
					"2023-06-07": {},
					"2023-06-06": {},
					"2023-06-05": {},
					"2023-06-04": {},
				},
			},
			nDays:    7,
			expected: 7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filteredResp, err := tc.apiResp.FilteredByDays(tc.nDays)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			actual := len(filteredResp.TimeSeriesData)
			if actual != tc.expected {
				t.Errorf("Expected %d days, got %d", tc.expected, actual)
			}
		})
	}
}
