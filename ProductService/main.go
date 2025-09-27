// cmd/product/main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type getProductRequest struct {
	Token string `json:"token"`
	SKU   uint64 `json:"sku"`
}
type getProductResponse struct {
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}

type repo struct {
	db *pgxpool.Pool

	wg sync.RWMutex
}

func (r *repo) GetBySKU(ctx context.Context, sku uint64) (*getProductResponse, error) {
	r.wg.Lock()
	defer r.wg.Unlock()

	const q = `SELECT name, price FROM Products WHERE sku = $1`
	var res getProductResponse
	if err := r.db.QueryRow(ctx, q, sku).Scan(&res.Name, &res.Price); err != nil {
		return nil, err
	}
	return &res, nil
}

func main() {
	var (
		addr  = flag.String("addr", getEnv("ADDR", ":8081"), "listen address")
		token = flag.String("token", getEnv("TOKEN", "dev-token"), "auth token to accept")
		dsn   = flag.String("dsn", getEnv("DATABASE_URL", "postgres://postgres:Verbov323213@localhost:5432/productDB"), "Postgres DSN")
	)
	flag.Parse()

	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(*dsn)
	if err != nil {
		log.Fatalf("parse dsn: %v", err)
	}
	cfg.MaxConns = 10
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	r := &repo{db: db}

	mux := http.NewServeMux()

	mux.HandleFunc("/get_product", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body getProductRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if body.Token != *token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if body.SKU == 0 {
			http.Error(w, "invalid sku", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		pr, err := r.GetBySKU(ctx, body.SKU)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, pr, http.StatusOK)
	})

	mux.HandleFunc("/product", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if req.URL.Query().Get("token") != *token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		skuStr := req.URL.Query().Get("sku")
		if skuStr == "" {
			http.Error(w, "missing sku", http.StatusBadRequest)
			return
		}
		sku, err := strconv.ParseUint(skuStr, 10, 64)
		if err != nil || sku == 0 {
			http.Error(w, "invalid sku", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		pr, err := r.GetBySKU(ctx, sku)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, pr, http.StatusOK)
	})

	log.Printf("ProductService listening on %s (token=%q)", *addr, *token)
	if err := http.ListenAndServe(*addr, mux); err != nil {
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
