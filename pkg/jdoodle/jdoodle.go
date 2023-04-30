package jdoodle

import "github.com/zekurio/daemon/pkg/httputils"

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

	res, err := httputils.Post(url, nil, payload)
	if err != nil {
		return
	}
	defer res.Release()

	err = res.JSON(result)

	return

}

// Credits returns the amount of credits left for the client
func (w *Wrapper) Credits() (result *CreditsResponse, err error) {
	payload := &credentialsBody{
		ClientID:     w.clientId,
		ClientSecret: w.clientSecret,
	}

	url := BaseURL + "/credits"
	result = &CreditsResponse{}

	res, err := httputils.Post(url, nil, payload)
	if err != nil {
		return
	}
	defer res.Release()

	err = res.JSON(result)

	return
}
