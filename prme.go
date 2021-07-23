package prme

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type client struct {
	token, apiHost string
	httpClient     *http.Client
}

// clientOption specifies prme client options as functions.
type clientOption func(*client) error

// WithAPIHost sets the Github API hostname for an instance of the client.
func WithAPIHost(host string) clientOption {
	return func(c *client) error {
		c.apiHost = host
		return nil
	}
}

// WithHTTPClient sets a custom net/http.Client for an instance of the client.
func WithHTTPClient(hc *http.Client) clientOption {
	return func(c *client) error {
		c.httpClient = hc
		return nil
	}
}

func New(token string, options ...clientOption) (*client, error) {
	if token == "" {
		return nil, fmt.Errorf("the Github token can not be an empty string, please specify a personal access token")
	}

	c := &client{
		token:   token,
		apiHost: "https://api.github.com",
	}

	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c client) GetRepoID(repo string) (int64, error) {
	apiURL := fmt.Sprintf("%s/repos/%s", c.apiHost, repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Github HTTP %s: %v", resp.Status, string(data))
	}

	var apiResp struct{ Id int64 }
	err = json.Unmarshal(data, &apiResp)
	if err != nil {
		return 0, err
	}

	return apiResp.Id, nil
}
