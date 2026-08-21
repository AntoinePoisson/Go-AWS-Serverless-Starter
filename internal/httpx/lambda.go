package httpx

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const readHeaderTimeout = 10 * time.Second

// Serve runs h as a Lambda function on Lambda, and as a plain HTTP server
// anywhere else.
func Serve(h http.Handler, addr string) error {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		slog.Info("listening", "addr", addr)
		server := &http.Server{
			Addr:              addr,
			Handler:           h,
			ReadHeaderTimeout: readHeaderTimeout,
		}
		return server.ListenAndServe()
	}

	lambda.Start(proxy(h))
	return nil
}

// proxy adapts h to the API Gateway HTTP API payload format 2.0.
func proxy(h http.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		req, err := newRequest(ctx, event)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, err
		}

		rec := &recorder{header: make(http.Header)}
		h.ServeHTTP(rec, req)

		return rec.response(), nil
	}
}

func newRequest(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	body := event.Body
	if event.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return nil, fmt.Errorf("decode request body: %w", err)
		}
		body = string(decoded)
	}

	host := event.RequestContext.DomainName
	if host == "" {
		host = "localhost"
	}

	target := event.RawPath
	if target == "" {
		target = "/"
	}
	if event.RawQueryString != "" {
		target += "?" + event.RawQueryString
	}

	req, err := http.NewRequestWithContext(ctx, event.RequestContext.HTTP.Method, "https://"+host+target, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	for name, value := range event.Headers {
		req.Header.Set(name, value)
	}
	for _, cookie := range event.Cookies {
		req.Header.Add("Cookie", cookie)
	}

	req.RemoteAddr = event.RequestContext.HTTP.SourceIP
	req.ContentLength = int64(len(body))

	return req, nil
}

// recorder captures what a handler writes, to turn it into a Lambda response.
type recorder struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func (r *recorder) Header() http.Header { return r.header }

// Write mirrors net/http and sniffs the content type when the handler set
// none. That also decides how the body is encoded.
func (r *recorder) Write(b []byte) (int, error) {
	if r.header.Get("Content-Type") == "" {
		r.header.Set("Content-Type", http.DetectContentType(b))
	}
	return r.body.Write(b)
}

func (r *recorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
}

func (r *recorder) response() events.APIGatewayV2HTTPResponse {
	headers := make(map[string]string, len(r.header))
	var cookies []string

	for name, values := range r.header {
		if strings.EqualFold(name, "Set-Cookie") {
			cookies = append(cookies, values...)
			continue
		}
		headers[name] = strings.Join(values, ", ")
	}

	status := r.status
	if status == 0 {
		status = http.StatusOK
	}

	body := r.body.String()
	encoded := r.body.Len() > 0 && !isTextual(r.header.Get("Content-Type"))
	if encoded {
		body = base64.StdEncoding.EncodeToString(r.body.Bytes())
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      status,
		Headers:         headers,
		Cookies:         cookies,
		Body:            body,
		IsBase64Encoded: encoded,
	}
}

// isTextual reports whether a content type survives as a plain string.
// Anything else has to reach API Gateway base64 encoded.
func isTextual(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	switch {
	case strings.HasPrefix(mediaType, "text/"),
		strings.HasSuffix(mediaType, "+json"),
		strings.HasSuffix(mediaType, "+xml"):
		return true
	}

	switch mediaType {
	case "application/json", "application/xml", "application/javascript", "application/x-www-form-urlencoded":
		return true
	default:
		return false
	}
}
