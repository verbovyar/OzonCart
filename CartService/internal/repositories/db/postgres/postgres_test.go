package postgres_test

import (
	"context"
	"errors"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
)

func TestAddItem_OK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`(?i)INSERT\s+INTO\s+Cart`).
		WithArgs(uint64(1), uint64(1001), uint64(2)).
		WillReturnRows(pgxmock.NewRows([]string{"dummy"}).AddRow(1))
	store := postgres.New(mockPool)
	err = store.AddItem(ctx, 1, 1001, 2)

	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAddItem_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	mockPool.ExpectQuery(`(?i)INSERT\s+INTO\s+Cart`).
		WithArgs(uint64(1), uint64(1001), uint64(2)).
		WillReturnError(errors.New("db fail"))

	store := postgres.New(mockPool)
	err := store.AddItem(ctx, 1, 1001, 2)

	require.Error(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestDeleteItem_OK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	mockPool.ExpectQuery(`(?i)DELETE\s+FROM\s+Cart\s+WHERE\s+user_id=\$1\s+AND\s+sku_id=\$2`).
		WithArgs(uint64(1), uint64(1001)).
		WillReturnRows(pgxmock.NewRows([]string{"dummy"}).AddRow(1))

	store := postgres.New(mockPool)
	err := store.DeleteItem(ctx, 1, 1001)

	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestClearCart_OK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	mockPool.ExpectQuery(`(?i)DELETE\s+FROM\s+Cart\s+WHERE\s+user_id=\$1`).
		WithArgs(uint64(7)).
		WillReturnRows(pgxmock.NewRows([]string{"dummy"}).AddRow(1))

	store := postgres.New(mockPool)
	err := store.ClearCart(ctx, 7)

	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestGetCart_OK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	rows := pgxmock.NewRows([]string{"sku_id", "count"}).
		AddRow(uint64(1001), uint64(2)).
		AddRow(uint64(1002), uint64(1))

	mockPool.ExpectQuery(`(?i)SELECT\s+sku_id,\s*count\s+FROM\s+Cart\s+WHERE\s+user_id=\$1\s+ORDER\s+BY\s+sku_id`).
		WithArgs(uint64(42)).
		WillReturnRows(rows)

	store := postgres.New(mockPool)
	out, err := store.GetCart(ctx, 42)

	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, uint64(1001), out[0].SkuID)
	require.Equal(t, uint64(2), out[0].Count)
	require.Equal(t, uint64(1002), out[1].SkuID)
	require.Equal(t, uint64(1), out[1].Count)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestGetCart_QueryError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	mockPool.ExpectQuery(`(?i)SELECT\s+sku_id,\s*count\s+FROM\s+Cart\s+WHERE\s+user_id=\$1\s+ORDER\s+BY\s+sku_id`).
		WithArgs(uint64(42)).
		WillReturnError(errors.New("query failed"))

	store := postgres.New(mockPool)
	_, err := store.GetCart(ctx, 42)

	require.Error(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestGetCart_ScanError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	rows := pgxmock.NewRows([]string{"sku_id", "count"}).
		AddRow("my", "bad")

	mockPool.ExpectQuery(`(?i)SELECT\s+sku_id,\s*count\s+FROM\s+Cart\s+WHERE\s+user_id=\$1\s+ORDER\s+BY\s+sku_id`).
		WithArgs(uint64(99)).
		WillReturnRows(rows)

	store := postgres.New(mockPool)
	_, err := store.GetCart(ctx, 99)

	require.Error(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}
