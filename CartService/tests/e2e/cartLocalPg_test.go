package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/verbovyar/OzonCart/internal/handlers"
	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
	"github.com/verbovyar/OzonCart/internal/service"
)

//go test -timeout 60s -run Test_E2E_LocalPostgres_And_FakeProductService ./tests/e2e -v

type getProductReq struct {
	Token string `json:"token"`
	SKU   uint64 `json:"sku"`
}

type getProductResp struct {
	Name  string `json:"name"`
	Price uint64 `json:"price"`
}

func startFakeProductService(t *testing.T, token string) *httptest.Server {
	t.Helper()

	products := map[uint64]getProductResp{
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
		var req getProductReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if req.Token != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		p, ok := products[req.SKU]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(p)
	})

	return httptest.NewServer(mux)
}

func newPgxPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := "postgres://postgres:Verbov323213@localhost:5432/ozonCartDB"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}

	return pool
}

func truncateCart(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `TRUNCATE TABLE Cart`)
	require.NoError(t, err)
}

func doPostJSON(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	resp, err := http.Post(url, "application/json", &buf)
	require.NoError(t, err)

	return resp
}

func doDelete(t *testing.T, url string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func doGet(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)

	return resp
}

func Test_E2E_LocalPostgres_And_FakeProductService(t *testing.T) {
	pool := newPgxPool(t)
	defer pool.Close()

	truncateCart(t, pool)

	const token = "dev-token"
	ps := startFakeProductService(t, token)
	defer ps.Close()

	repo := postgres.New(pool)
	pc := service.NewClient(ps.URL, token, 1, 50*time.Millisecond)
	cs := service.New(repo, pc)

	mux := http.NewServeMux()
	mux.Handle("/user/", handlers.New(cs))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	user := 1

	// add 2×1001, add 1×1002 -> get -> total=3900
	// delete 1002 -> get -> total=3000
	// clear -> get -> пусто

	resp := doPostJSON(t, fmt.Sprintf("%s/user/%d/cart/1001", ts.URL, user), map[string]any{"count": 2})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	resp = doPostJSON(t, fmt.Sprintf("%s/user/%d/cart/1002", ts.URL, user), map[string]any{"count": 1})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	type cartResp struct {
		Items []struct {
			SkuID uint64 `json:"sku_id"`
			Name  string `json:"name"`
			Count uint64 `json:"count"`
			Price uint64 `json:"price"`
		} `json:"items"`
		TotalPrice uint64 `json:"total_price"`
	}

	log.Println("Test: doGet")

	// get -> 2 позиции, total=3900
	resp = doGet(t, fmt.Sprintf("%s/user/%d/cart", ts.URL, user))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body cartResp
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	resp.Body.Close()
	require.Len(t, body.Items, 2)
	require.Equal(t, uint64(3900), body.TotalPrice)

	log.Println("Test: doDelete1")
	// delete 1002
	resp = doDelete(t, fmt.Sprintf("%s/user/%d/cart/1002", ts.URL, user))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	log.Println("Test: doGet1")
	// get -> остался только 1001, total=3000
	resp = doGet(t, fmt.Sprintf("%s/user/%d/cart", ts.URL, user))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	resp.Body.Close()
	require.Len(t, body.Items, 1)
	require.Equal(t, uint64(3000), body.TotalPrice)

	log.Println("Test: doDelete2")
	// clear
	resp = doDelete(t, fmt.Sprintf("%s/user/%d/cart", ts.URL, user))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	log.Println("Test: doGet2")
	// get -> пусто
	resp = doGet(t, fmt.Sprintf("%s/user/%d/cart", ts.URL, user))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	resp.Body.Close()
	require.Equal(t, 0, len(body.Items))
}
