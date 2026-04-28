package models

type QueryResponse struct {
	Data []map[string]any `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
