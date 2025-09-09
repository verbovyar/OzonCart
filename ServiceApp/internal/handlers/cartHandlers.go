package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/verbovyar/OzonCart/internal/domain"
	"github.com/verbovyar/OzonCart/internal/service"
	"github.com/verbovyar/OzonCart/internal/validation"
)

const BASE = 10
const TYPE = 64
const URL_PARTS_COUNT = 3

var ErrBadId = errors.New("bad id")

type CartRouter struct {
	cs *service.CartService
	v  *validation.Validator
}

func New(cs *service.CartService) http.Handler {
	r := &CartRouter{
		cs: cs,
		v:  validation.New(),
	}

	return http.HandlerFunc(r.root)
}

func (c *CartRouter) root(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(parts) < URL_PARTS_COUNT || parts[0] != "user" || parts[2] != "cart" {
		http.NotFound(w, req)
		return
	}

	userID, err := parseID(parts[1])
	if err != nil || !c.v.ValidateID(userID) {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodPost:
		if len(parts) != 4 {
			http.NotFound(w, req)
			return
		}
		c.addToCart(w, req, userID, parts[3])
	case http.MethodDelete:
		if len(parts) == 4 {
			c.deleteItem(w, req, userID, parts[3])
		} else if len(parts) == 3 {
			c.clearCart(w, req, userID)
		} else {
			http.NotFound(w, req)
			return
		}
	case http.MethodGet:
		if len(parts) != 3 {
			http.NotFound(w, req)
			return
		}
		c.getCart(w, req, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func parseID(s string) (uint64, error) {
	id, err := strconv.ParseInt(s, BASE, TYPE)
	if err != nil || id <= 0 {
		return 0, ErrBadId
	}

	return uint64(id), nil
}

func (c *CartRouter) addToCart(w http.ResponseWriter, req *http.Request, userID uint64, skuStr string) {
	skuID, err := parseID(skuStr)
	if err != nil || !c.v.ValidateID(skuID) {
		http.Error(w, "Invalid sku_id", http.StatusBadRequest)
		return
	}

	var body domain.AddToCartRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if err := c.v.Struct(body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.cs.AddToCart(req.Context(), userID, skuID, body.Count); err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			http.Error(w, "Product not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *CartRouter) deleteItem(w http.ResponseWriter, _ *http.Request, userID uint64, skuStr string) {
	skuID, err := parseID(skuStr)
	if err != nil {
		http.Error(w, "Invalid sku_id", http.StatusBadRequest)
		return
	}
	c.cs.DeleteItem(userID, skuID)

	w.WriteHeader(http.StatusNoContent)
}

func (c *CartRouter) clearCart(w http.ResponseWriter, _ *http.Request, userID uint64) {
	c.cs.ClearCart(userID)

	w.WriteHeader(http.StatusNoContent)
}

func (c *CartRouter) getCart(w http.ResponseWriter, req *http.Request, userID uint64) {
	resp, err := c.cs.GetCart(req.Context(), userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if len(resp.Items) == 0 {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
