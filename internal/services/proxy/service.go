package proxy

import (
	"net/http"
)

// Service is a placeholder that will route chat/completion requests to providers.
type Service struct{}

func NewService() Service { return Service{} }

// ProxyChatCompletions is a stub that we'll implement later.
func (s Service) ProxyChatCompletions(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotImplemented)
    _, _ = w.Write([]byte("proxy not implemented yet"))
}


