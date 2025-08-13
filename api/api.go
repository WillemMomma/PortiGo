package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type ModelParams struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Endpoint string `json:"endpoint"`
}

type ModelResponse struct {
	Data []ModelParams `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (m *ModelParams) Create(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ModelResponse{Data: []ModelParams{*m}})
}