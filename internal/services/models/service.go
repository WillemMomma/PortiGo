package models

import (
	"context"
	"net/http"

	"go-gateway/internal/domain/model"
)

// Repository abstracts persistence for models.
type Repository interface {
	List(ctx context.Context) ([]model.Model, error)
	Create(ctx context.Context, in model.CreateModelInput) (model.Model, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service { return Service{repo: repo} }

// The following methods implement httpx.ModelService interface without importing it.
func (s Service) ListModels(r *http.Request) ([]model.Model, error) {
	return s.repo.List(r.Context())
}

func (s Service) CreateModel(r *http.Request, in model.CreateModelInput) (model.Model, error) {
	return s.repo.Create(r.Context(), in)
}


