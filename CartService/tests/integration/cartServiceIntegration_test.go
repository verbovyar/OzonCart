package integration_test

import (
	"context"
	"testing"
	"time"

	pgstore "github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
	"github.com/verbovyar/OzonCart/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func mustNewPGPool(t *testing.T) *pgxpool.Pool {
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
	_, err := pool.Exec(context.Background(), `TRUNCATE TABLE Cart RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

type fakeProduct struct {
	Name  string
	Price uint64
}

type fakeProductClient struct {
	bySKU map[uint64]fakeProduct
}

func (f *fakeProductClient) GetProduct(_ context.Context, skuID uint64) (*service.Product, error) {
	if p, ok := f.bySKU[skuID]; ok {
		return &service.Product{Name: p.Name, Price: p.Price}, nil
	}

	return nil, service.ErrProductNotFound
}

func Test_CartService_Integration_With_LocalPostgres(t *testing.T) {
	pool := mustNewPGPool(t)
	defer pool.Close()
	truncateCart(t, pool)

	repo := pgstore.New(pool)

	pc := &fakeProductClient{
		bySKU: map[uint64]fakeProduct{
			1001: {Name: "Demo T-Shirt", Price: 1500},
			1002: {Name: "Coffee Mug", Price: 900},
			1003: {Name: "Sticker Pack", Price: 300},
		},
	}

	cs := service.New(repo, pc)

	ctx := context.Background()
	var (
		userID uint64 = 1
	)

	// add 2 × 1001
	require.NoError(t, cs.AddToCart(ctx, userID, 1001, 2))
	// add 1 × 1002
	require.NoError(t, cs.AddToCart(ctx, userID, 1002, 1))

	// get -> 2 позиции, total = 2*1500 + 1*900 = 3900
	res, err := cs.GetCart(ctx, userID)
	require.NoError(t, err)
	require.Len(t, res.Items, 2)
	require.Equal(t, uint64(3900), res.TotalPrice)

	// delete 1002 -> остается только 1001 (2×1500 = 3000)
	cs.DeleteItem(userID, 1002)
	res, err = cs.GetCart(ctx, userID)
	require.NoError(t, err)
	require.Len(t, res.Items, 1)
	require.Equal(t, uint64(3000), res.TotalPrice)

	// clear -> пустая корзина, 200-ок в HTTP, здесь просто данные = 0
	cs.ClearCart(userID)
	res, err = cs.GetCart(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 0, len(res.Items))
	require.Equal(t, uint64(0), res.TotalPrice)
}
