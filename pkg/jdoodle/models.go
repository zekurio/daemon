package jdoodle

type credentialsBody struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type ExecRequestBody struct {
	*credentialsBody

	Script   string `json:"script"`
	Language string `json:"language"`
}

type ExecResponse struct {
	Output            string `json:"output"`
	Memory            string `json:"memory"`
	CPUTime           string `json:"cpuTime"`
	Status            string `json:"statusCode"`
	CompilationStatus string `json:"compilationStatus"`
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Status string `json:"statusCode"`
}

type CreditsResponse struct {
	Used int `json:"used"`
}
