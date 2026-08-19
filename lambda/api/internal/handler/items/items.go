// Package items exposes the CRUD routes of the item resource.
package items

import (
	"net/http"
	"strconv"

	"github.com/google/wire"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
)

// WireSet provides the handler.
var WireSet = wire.NewSet(New)

// Handler serves the item resource.
type Handler struct {
	items item.ServiceAPI
}

// New returns a Handler backed by the given service.
func New(items item.ServiceAPI) *Handler {
	return &Handler{items: items}
}

// AddRoutes registers the item routes.
func (h *Handler) AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /items", httpx.Handle(h.create))
	mux.HandleFunc("GET /items", httpx.Handle(h.list))
	mux.HandleFunc("GET /items/{id}", httpx.Handle(h.get))
	mux.HandleFunc("DELETE /items/{id}", httpx.Handle(h.delete))
}

// ListResponse is the body returned by GET /items. Exported so the generated
// specification names the schema after it.
type ListResponse struct {
	Items []item.Item `json:"items"`
	Count int         `json:"count"`
}

// create stores a new item.
//
//	@Id				createItem
//	@Summary		Create an item
//	@Description	Stores an item and returns it with its generated identifier.
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Param			body	body		item.CreateInput	true	"Item to create"
//	@Success		201		{object}	item.Item
//	@Header			201		{string}	Location	"Path of the created item"
//	@Failure		400		{object}	httpx.Error	"invalid_body or invalid_input"
//	@Failure		401		{object}	httpx.Error	"unauthorized"
//	@Failure		413		{object}	httpx.Error	"body_too_large, over 1 MiB"
//	@Failure		500		{object}	httpx.Error	"internal_error"
//	@Security		ApiKeyAuth
//	@Router			/items [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var input item.CreateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		return err
	}

	created, err := h.items.Create(r.Context(), input)
	if err != nil {
		return item.HTTPError(err)
	}

	w.Header().Set("Location", "/items/"+created.ID)
	httpx.WriteJSON(w, http.StatusCreated, created)
	return nil
}

// get returns a single item.
//
//	@Id				getItem
//	@Summary		Read an item
//	@Description	Any non-empty identifier is accepted; an unknown one is a 404.
//	@Tags			items
//	@Produce		json
//	@Param			id	path		string	true	"Item identifier"
//	@Success		200	{object}	item.Item
//	@Failure		401	{object}	httpx.Error	"unauthorized"
//	@Failure		404	{object}	httpx.Error	"not_found"
//	@Failure		500	{object}	httpx.Error	"internal_error"
//	@Security		ApiKeyAuth
//	@Router			/items/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	found, err := h.items.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return item.HTTPError(err)
	}

	httpx.WriteJSON(w, http.StatusOK, found)
	return nil
}

// delete removes an item.
//
//	@Id				deleteItem
//	@Summary		Delete an item
//	@Description	Deleting an identifier that does not exist is a 404, not a no-op.
//	@Tags			items
//	@Produce		json
//	@Param			id	path	string	true	"Item identifier"
//	@Success		204	"Deleted, no body"
//	@Failure		401	{object}	httpx.Error	"unauthorized"
//	@Failure		404	{object}	httpx.Error	"not_found"
//	@Failure		500	{object}	httpx.Error	"internal_error"
//	@Security		ApiKeyAuth
//	@Router			/items/{id} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	if err := h.items.Delete(r.Context(), r.PathValue("id")); err != nil {
		return item.HTTPError(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

// list returns a page of items.
//
//	@Id				listItems
//	@Summary		List items
//	@Description	A limit outside [1, 100] falls back to 25. The count is the size of
//	@Description	the page, not a total: there is no cursor and no total count.
//	@Tags			items
//	@Produce		json
//	@Param			limit	query		int	false	"Page size, 1 to 100"	default(25)
//	@Success		200		{object}	ListResponse
//	@Failure		400		{object}	httpx.Error	"invalid_limit"
//	@Failure		401		{object}	httpx.Error	"unauthorized"
//	@Failure		500		{object}	httpx.Error	"internal_error"
//	@Security		ApiKeyAuth
//	@Router			/items [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	var limit int32
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return httpx.Errorf(http.StatusBadRequest, "invalid_limit", "limit must be an integer")
		}
		limit = int32(parsed)
	}

	found, err := h.items.List(r.Context(), limit)
	if err != nil {
		return item.HTTPError(err)
	}
	if found == nil {
		found = []item.Item{}
	}

	httpx.WriteJSON(w, http.StatusOK, ListResponse{Items: found, Count: len(found)})
	return nil
}
