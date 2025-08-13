package postgres

import (
	"context"
	"database/sql"
	"errors"

	"go-gateway/internal/domain/model"
)

type ModelRepository struct {
	db *sql.DB
}

func NewModelRepository(db *sql.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

// EnsureSchema creates the minimal tables if they do not exist.
func (r *ModelRepository) EnsureSchema(ctx context.Context) error {
	// Minimal schema: models with id, name, description, endpoint, api_key
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS models (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			endpoint TEXT NOT NULL,
			api_key TEXT NOT NULL
		);
	`)
	return err
}

func (r *ModelRepository) List(ctx context.Context) ([]model.Model, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description, endpoint FROM models ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Model
	for rows.Next() {
		var m model.Model
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Endpoint); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ModelRepository) Create(ctx context.Context, in model.CreateModelInput) (model.Model, error) {
    if in.ID == "" || in.Name == "" || in.Endpoint == "" {
        return model.Model{}, errors.New("id, name, endpoint are required")
    }
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO models (id, name, description, endpoint, api_key)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			endpoint = EXCLUDED.endpoint,
			api_key = EXCLUDED.api_key
	`, in.ID, in.Name, in.Description, in.Endpoint, in.APIKey)
	if err != nil {
		return model.Model{}, err
	}
	return model.Model{ID: in.ID, Name: in.Name, Description: in.Description, Endpoint: in.Endpoint}, nil
}

// GetByID fetches a model by id including its stored api_key.
func (r *ModelRepository) GetByID(ctx context.Context, id string) (model.Model, string, error) {
    var m model.Model
    var apiKey string
    row := r.db.QueryRowContext(ctx, `SELECT id, name, description, endpoint, api_key FROM models WHERE id = $1`, id)
    if err := row.Scan(&m.ID, &m.Name, &m.Description, &m.Endpoint, &apiKey); err != nil {
        return model.Model{}, "", err
    }
    return m, apiKey, nil
}


