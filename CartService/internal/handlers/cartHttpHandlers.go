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

type CartHttpRouter struct {
	cs *service.CartService
	v  *validation.Validator
}

func New(cs *service.CartService) http.Handler {
	r := &CartHttpRouter{
		cs: cs,
		v:  validation.New(),
	}

	return http.HandlerFunc(r.root)
}

func (c *CartHttpRouter) root(w http.ResponseWriter, req *http.Request) {
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

// addToCart godoc
// @Summary      Добавить товар в корзину
// @Description  Добавляет SKU в корзину пользователя после проверки существования во внешнем ProductService
// @Tags         cart
// @Accept       json
// @Param        user_id path int true "ID пользователя"
// @Param        sku_id  path int true "SKU товара"
// @Param        payload body domain.AddToCartRequest true "Количество"
// @Success      200 "OK"
// @Failure      400 {string} string "invalid input"
// @Failure      404 {string} string "product not found"
// @Failure      500 {string} string "server error"
// @Router       /user/{user_id}/cart/{sku_id} [post]
func (c *CartHttpRouter) addToCart(w http.ResponseWriter, req *http.Request, userID uint64, skuStr string) {
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

// deleteItem godoc
// @Summary      Удалить товар из корзины
// @Tags         cart
// @Param        user_id path int true "ID пользователя"
// @Param        sku_id  path int true "SKU"
// @Success      204 "No Content"
// @Router       /user/{user_id}/cart/{sku_id} [delete]
func (c *CartHttpRouter) deleteItem(w http.ResponseWriter, _ *http.Request, userID uint64, skuStr string) {
	skuID, err := parseID(skuStr)
	if err != nil {
		http.Error(w, "Invalid sku_id", http.StatusBadRequest)
		return
	}
	c.cs.DeleteItem(userID, skuID)

	w.WriteHeader(http.StatusNoContent)
}

// clearCart godoc
// @Summary      Очистить корзину пользователя
// @Tags         cart
// @Param        user_id path int true "ID пользователя"
// @Success      204 "No Content"
// @Router       /user/{user_id}/cart [delete]
func (c *CartHttpRouter) clearCart(w http.ResponseWriter, _ *http.Request, userID uint64) {
	c.cs.ClearCart(userID)

	w.WriteHeader(http.StatusNoContent)
}

// getCart godoc
// @Summary      Получить содержимое корзины
// @Tags         cart
// @Param        user_id path int true "ID пользователя"
// @Success      200 {object} domain.GetCartResponse
// @Failure      404 {string} string "cart is empty"
// @Router       /user/{user_id}/cart [get]
func (c *CartHttpRouter) getCart(w http.ResponseWriter, req *http.Request, userID uint64) {
	resp, err := c.cs.GetCart(req.Context(), userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
