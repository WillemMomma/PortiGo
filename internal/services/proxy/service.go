package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	dmodel "go-gateway/internal/domain/model"
)

// ModelLookup abstracts looking up a model by id.
type ModelLookup interface {
    GetByID(ctx context.Context, id string) (dmodel.Model, string, error)
}

// Service proxies OpenAI-compatible requests to the provider endpoint.
type Service struct {
    repo   ModelLookup
    client *http.Client
}

func NewService(repo ModelLookup) Service {
    return Service{
        repo: repo,
        client: &http.Client{Timeout: 0}, // no overall timeout to allow streaming
    }
}

// ProxyChatCompletions forwards the incoming request to the model's endpoint selected by body.model.
func (s Service) ProxyChatCompletions(w http.ResponseWriter, r *http.Request) {
    // Read body to extract the model id while keeping the original body for forwarding
    var buf bytes.Buffer
    if _, err := io.Copy(&buf, r.Body); err != nil {
        http.Error(w, "failed to read request body", http.StatusBadRequest)
        return
    }
    _ = r.Body.Close()

    modelID, err := extractModelID(buf.Bytes())
    if err != nil || modelID == "" {
        http.Error(w, "missing model in request body", http.StatusBadRequest)
        return
    }

    // Lookup model to get endpoint
    mdl, apiKey, err := s.repo.GetByID(r.Context(), modelID)
    if err != nil {
        http.Error(w, "unknown model id", http.StatusNotFound)
        return
    }

    targetURL, err := joinEndpointAndPath(mdl.Endpoint, r.URL)
    if err != nil {
        http.Error(w, "invalid endpoint", http.StatusBadGateway)
        return
    }

    // Create outbound request with original headers/body
    outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(buf.Bytes()))
    if err != nil {
        http.Error(w, "failed to create upstream request", http.StatusBadGateway)
        return
    }
    copyHeaders(outReq.Header, r.Header)
    // Ensure content-type is preserved
    if ct := r.Header.Get("Content-Type"); ct != "" {
        outReq.Header.Set("Content-Type", ct)
    }
    // If upstream needs auth and incoming request doesn't set it, we can set from stored apiKey
    if outReq.Header.Get("Authorization") == "" && strings.TrimSpace(apiKey) != "" {
        outReq.Header.Set("Authorization", "Bearer "+apiKey)
    }

    // Perform request
    resp, err := s.client.Do(outReq)
    if err != nil {
        http.Error(w, "upstream request failed", http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    // Copy response headers (excluding hop-by-hop) and status
    copyResponseHeaders(w.Header(), resp.Header)
    w.WriteHeader(resp.StatusCode)

    // Stream body to client
    flusher, _ := w.(http.Flusher)
    copyStreaming(w, resp.Body, flusher)
}

// extractModelID parses a minimal JSON body to find the "model" field.
func extractModelID(body []byte) (string, error) {
    // minimal, allocation-friendly search before falling back to JSON decode
    // This handles common bodies where "model":"..." appears.
    lower := strings.ToLower(string(body))
    if !strings.Contains(lower, "\"model\"") {
        return "", errors.New("model not present")
    }
    // Fallback to robust JSON decode
    type onlyModel struct{ Model string `json:"model"` }
    var m onlyModel
    // Use a tiny decoder to avoid re-escaping
    dec := jsonNewDecoder(bytes.NewReader(body))
    if err := dec.Decode(&m); err == nil && m.Model != "" {
        return m.Model, nil
    }
    // If body is not a single JSON object (e.g., contains whitespace or BOM), try again strictly
    dec = jsonNewDecoder(bytes.NewReader(body))
    if err := dec.Decode(&m); err != nil {
        return "", err
    }
    return m.Model, nil
}

// joinEndpointAndPath ensures endpoint base and incoming path/query are combined correctly.
func joinEndpointAndPath(endpoint string, in *url.URL) (string, error) {
    if endpoint == "" {
        return "", errors.New("empty endpoint")
    }
    base := strings.TrimRight(endpoint, "/")
    path := in.Path
    if !strings.HasPrefix(path, "/") {
        path = "/" + path
    }
    if in.RawQuery != "" {
        return base + path + "?" + in.RawQuery, nil
    }
    return base + path, nil
}

// copyHeaders copies request headers except hop-by-hop ones.
func copyHeaders(dst, src http.Header) {
    for k, vv := range src {
        if isHopByHopHeader(k) {
            continue
        }
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}

// copyResponseHeaders copies response headers excluding hop-by-hop ones.
func copyResponseHeaders(dst, src http.Header) {
    for k, vv := range src {
        if isHopByHopHeader(k) {
            continue
        }
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}

func isHopByHopHeader(h string) bool {
    switch textproto.CanonicalMIMEHeaderKey(h) {
    case "Connection", "Proxy-Connection", "Keep-Alive", "TE", "Trailer", "Transfer-Encoding", "Upgrade":
        return true
    default:
        return false
    }
}

func copyStreaming(dst io.Writer, src io.Reader, flusher http.Flusher) {
    buf := make([]byte, 32*1024)
    for {
        n, err := src.Read(buf)
        if n > 0 {
            _, _ = dst.Write(buf[:n])
            if flusher != nil {
                flusher.Flush()
            }
        }
        if err != nil {
            if errors.Is(err, io.EOF) {
                return
            }
            return
        }
    }
}

// Minimal JSON decoder wrapper to avoid importing encoding/json in many places
// while keeping code readable and testable.
type jsonDecoder interface{ Decode(v any) error }

func jsonNewDecoder(r io.Reader) jsonDecoder { return &stdJSONDecoder{dec: json.NewDecoder(r)} }

type stdJSONDecoder struct{ dec *json.Decoder }

func (d *stdJSONDecoder) Decode(v any) error { return d.dec.Decode(v) }


