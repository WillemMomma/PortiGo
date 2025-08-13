package model

// Model represents a minimal model record for routing.
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
}

// CreateModelInput represents input for creating a model.
type CreateModelInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	APIKey      string `json:"api_key"`
}


