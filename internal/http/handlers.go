package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"

	dmodel "go-gateway/internal/domain/model"
)

// Handler exposes HTTP endpoints; dependencies injected via interfaces.
type Handler struct {
    Models   ModelService
    Proxy    ProxyService
}

// ModelService represents CRUD of models in storage.
type ModelService interface {
    ListModels(r *http.Request) ([]dmodel.Model, error)
    CreateModel(r *http.Request, m dmodel.CreateModelInput) (dmodel.Model, error)
}

// ProxyService proxies chat/completion requests to a provider.
type ProxyService interface {
    ProxyChatCompletions(w http.ResponseWriter, r *http.Request)
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
}

func (h Handler) ListModels(w http.ResponseWriter, r *http.Request) {
    models, err := h.Models.ListModels(r)
    if err != nil {
        slog.Error("list models", slog.Any("err", err))
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list models"})
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"data": models})
}

func (h Handler) CreateModel(w http.ResponseWriter, r *http.Request) {
    var in dmodel.CreateModelInput
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    m, err := h.Models.CreateModel(r, in)
    if err != nil {
        slog.Error("create model", slog.Any("err", err))
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusCreated, map[string]any{"data": m})
}

func (h Handler) ProxyChatCompletions(w http.ResponseWriter, r *http.Request) {
    h.Proxy.ProxyChatCompletions(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}


