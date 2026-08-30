// Package health exposes the liveness endpoint of the service.
package health

import (
	"net/http"

	"github.com/google/wire"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

// WireSet provides the handler.
var WireSet = wire.NewSet(New)

// Handler serves the health endpoint.
type Handler struct {
	stage   string
	version string
}

// New returns a Handler reporting the running stage and version.
func New(cfg *config.Config) *Handler {
	return &Handler{stage: cfg.Stage, version: cfg.Version}
}

// AddRoutes registers the health route.
func (h *Handler) AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.get)
}

// Response is the body of GET /health. Exported so the generated schema is
// named after it.
type Response struct {
	Status  string `json:"status" validate:"required" example:"ok"`
	Stage   string `json:"stage" validate:"required" example:"prod"`
	Version string `json:"version" validate:"required" example:"1.2.3+abc1234"`
}

// get reports that the function is alive, and which build answered.
//
//	@Id				getHealth
//	@Summary		Liveness, stage and version
//	@Description	Answers 200 whenever the function runs. The version identifies the
//	@Description	deployed build; it reads "dev" outside a deployed stage.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	Response
//	@Router			/health [get]
func (h *Handler) get(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, Response{
		Status:  "ok",
		Stage:   h.stage,
		Version: h.version,
	})
}
