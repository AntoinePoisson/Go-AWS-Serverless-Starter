// Package items is the read-only item routes.
package items

import (
	"net/http"

	"github.com/google/wire"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
)

// WireSet is the handler.
var WireSet = wire.NewSet(New)

// Handler is the public item reads.
type Handler struct {
	items item.ServiceAPI
}

// New wraps the item service.
func New(items item.ServiceAPI) *Handler {
	return &Handler{items: items}
}

// AddRoutes registers the public reads. Under /public so they dont collide
// with the api function on the same gateway.
func (h *Handler) AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /public/items/{id}", httpx.Handle(h.get))
}

// get reads one item, no key.
//
//	@Id				getPublicItem
//	@Summary		Read an item, unauthenticated
//	@Description	Same payload as GET /items/{id}, without the API key.
//	@Tags			public
//	@Produce		json
//	@Param			id	path		string	true	"Item identifier"
//	@Success		200	{object}	item.Item
//	@Failure		404	{object}	httpx.Error	"not_found"
//	@Failure		500	{object}	httpx.Error	"internal_error"
//	@Router			/public/items/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	found, err := h.items.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return item.HTTPError(err)
	}

	httpx.WriteJSON(w, http.StatusOK, found)
	return nil
}
