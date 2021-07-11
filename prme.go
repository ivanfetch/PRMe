package prme

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type client struct {
	token string
}

func New(token string) (*client, error) {
	return &client{
		token: token,
	}, nil
}

func (c client) GetRepoID(repo string) (int64, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))

	hc := http.Client{}
	resp, err := hc.Do(req)
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
