package jdoodle

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	// BaseURL is the base url for the jdoodle api
	BaseURL = "https://api.jdoodle.com/v1"
)

type Wrapper struct {
	clientId     string
	clientSecret string
}

func New(clientId, clientSecret string) *Wrapper {
	return &Wrapper{clientId, clientSecret}
}

// Execute uses the execute endpoint to run code, takes a language and script string
// and returns the result
func (w *Wrapper) Execute(language, script string) (result *ExecResponse, err error) {
	payload := &ExecRequestBody{
		credentialsBody: &credentialsBody{
			ClientID:     w.clientId,
			ClientSecret: w.clientSecret,
		},
		Script:   script,
		Language: language,
	}

	url := BaseURL + "/execute"
	result = &ExecResponse{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("execute request failed with status code " + resp.Status)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Credits returns the amount of credits left for the client
func (w *Wrapper) Credits() (result *CreditsResponse, err error) {
	payload := &credentialsBody{
		ClientID:     w.clientId,
		ClientSecret: w.clientSecret,
	}

	url := BaseURL + "/credits"
	result = &CreditsResponse{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("credits request failed with status code " + resp.Status)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
