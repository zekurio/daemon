package models

var (
	Ok = &Status{200}
)

type Error struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Context string `json:"context,omitempty"`
}

type Status struct {
	Code int `json:"code"`
}
