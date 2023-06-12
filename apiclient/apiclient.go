package apiclient

import (
	"net/http"
)

type APIClient struct {
	client *http.Client
	APIURL string
}

func NewAPIClient(client *http.Client, url string) *APIClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &APIClient{client: client, APIURL: url}
}

func (api *APIClient) DoAPIRequest(req *http.Request) (*http.Response, error) {
	return api.client.Do(req)
}
