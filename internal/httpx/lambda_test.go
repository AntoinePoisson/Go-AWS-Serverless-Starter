package httpx

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEvent(method, path string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		RawPath: path,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			DomainName: "api.example.com",
			HTTP:       events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: method, SourceIP: "203.0.113.7"},
		},
	}
}

func TestProxyRoutesRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{
			"id":    r.PathValue("id"),
			"query": r.URL.Query().Get("verbose"),
		})
	})

	event := newEvent(http.MethodGet, "/items/42")
	event.RawQueryString = "verbose=true"

	res, err := proxy(mux)(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, `{"id":"42","query":"true"}`, res.Body)
	assert.Equal(t, "application/json", res.Headers["Content-Type"])
}

func TestProxyForwardsHeadersAndBody(t *testing.T) {
	var (
		gotBody   string
		gotHeader string
		gotIP     string
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		gotHeader = r.Header.Get("X-Api-Key")
		gotIP = r.RemoteAddr
		w.WriteHeader(http.StatusCreated)
	})

	event := newEvent(http.MethodPost, "/items")
	event.Headers = map[string]string{"X-Api-Key": "secret"}
	event.Body = `{"name":"demo"}`

	res, err := proxy(mux)(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, `{"name":"demo"}`, gotBody)
	assert.Equal(t, "secret", gotHeader)
	assert.Equal(t, "203.0.113.7", gotIP)
}

func TestProxyDecodesBase64Body(t *testing.T) {
	var gotBody string

	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", func(_ http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
	})

	event := newEvent(http.MethodPost, "/items")
	event.Body = base64.StdEncoding.EncodeToString([]byte(`{"name":"demo"}`))
	event.IsBase64Encoded = true

	_, err := proxy(mux)(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, `{"name":"demo"}`, gotBody)
}

func TestProxyRejectsInvalidBase64Body(t *testing.T) {
	event := newEvent(http.MethodPost, "/items")
	event.Body = "not base64!"
	event.IsBase64Encoded = true

	_, err := proxy(http.NewServeMux())(context.Background(), event)
	assert.Error(t, err)
}

func TestProxyDefaultsToStatusOK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	res, err := proxy(mux)(context.Background(), newEvent(http.MethodGet, "/ping"))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "pong", res.Body)
	assert.False(t, res.IsBase64Encoded)
	assert.Contains(t, res.Headers["Content-Type"], "text/plain", "the content type must be inferred like net/http does")
}

func TestProxyEncodesBinaryResponses(t *testing.T) {
	payload := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /logo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(payload)
	})

	res, err := proxy(mux)(context.Background(), newEvent(http.MethodGet, "/logo"))
	require.NoError(t, err)

	assert.True(t, res.IsBase64Encoded)
	decoded, err := base64.StdEncoding.DecodeString(res.Body)
	require.NoError(t, err)
	assert.Equal(t, payload, decoded)
}

func TestProxyLeavesTextualResponsesAlone(t *testing.T) {
	for _, contentType := range []string{"application/json", "text/plain; charset=utf-8", "application/problem+json"} {
		t.Run(contentType, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", contentType)
				_, _ = w.Write([]byte("payload"))
			})

			res, err := proxy(mux)(context.Background(), newEvent(http.MethodGet, "/ping"))
			require.NoError(t, err)

			assert.False(t, res.IsBase64Encoded)
			assert.Equal(t, "payload", res.Body)
		})
	}
}

func TestProxySplitsSetCookieHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		w.WriteHeader(http.StatusOK)
	})

	res, err := proxy(mux)(context.Background(), newEvent(http.MethodGet, "/ping"))
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"a=1", "b=2"}, res.Cookies)
	assert.NotContains(t, res.Headers, "Set-Cookie")
}
