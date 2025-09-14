// main.go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
)

type getProductRequest struct {
	Token string `json:"token"`
	SKU   uint64 `json:"sku"`
}

type getProductResponse struct {
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}

func main() {
	var (
		addrFlag  = flag.String("addr", getEnv("ADDR", ":8081"), "listen address")
		tokenFlag = flag.String("token", getEnv("TOKEN", "dev-token"), "auth token to accept")
	)
	flag.Parse()

	products := map[uint64]getProductResponse{
		1001: {Name: "Demo T-Shirt", Price: 1500},
		1002: {Name: "Coffee Mug", Price: 900},
		1003: {Name: "Sticker Pack", Price: 300},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/get_product", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req getProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		if req.Token != *tokenFlag {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if req.SKU <= 0 {
			http.Error(w, "invalid sku", http.StatusBadRequest)
			return
		}

		pr, ok := products[req.SKU]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		writeJSON(w, pr, http.StatusOK)
	})

	mux.HandleFunc("/product", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		token := r.URL.Query().Get("token")
		if token != *tokenFlag {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		skuStr := r.URL.Query().Get("sku")
		if skuStr == "" {
			http.Error(w, "missing sku", http.StatusBadRequest)
			return
		}
		sku, err := strconv.ParseInt(skuStr, 10, 64)
		if err != nil || sku <= 0 {
			http.Error(w, "invalid sku", http.StatusBadRequest)
			return
		}
		switch sku {
		case 1001:
			writeJSON(w, getProductResponse{Name: "Demo T-Shirt", Price: 1500}, http.StatusOK)
		case 1002:
			writeJSON(w, getProductResponse{Name: "Coffee Mug", Price: 900}, http.StatusOK)
		case 1003:
			writeJSON(w, getProductResponse{Name: "Sticker Pack", Price: 300}, http.StatusOK)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	log.Printf("ProductService listening on %s (token=%q)", *addrFlag, *tokenFlag)
	if err := http.ListenAndServe(*addrFlag, mux); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
