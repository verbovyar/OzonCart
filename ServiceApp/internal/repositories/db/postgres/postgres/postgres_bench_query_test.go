package postgres_test

import (
	"context"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	pgstore "github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
)

func BenchmarkStore_AddItem_QueryStyle(b *testing.B) {
	ctx := context.Background()
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	store := pgstore.New(mockPool)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockPool.ExpectQuery(`(?i)^INSERT\s+INTO\s+cart`).
			WithArgs(uint64(1), uint64(10000+i), uint64(1)).
			WillReturnRows(pgxmock.NewRows([]string{"dummy"}).AddRow(1))

		store.AddItem(ctx, 1, uint64(10000+i), 1)
	}
	mockPool.ExpectationsWereMet()
}

func BenchmarkStore_DeleteItem_QueryStyle(b *testing.B) {
	ctx := context.Background()
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	store := pgstore.New(mockPool)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockPool.ExpectQuery(`(?i)^DELETE\s+FROM\s+cart\s+WHERE\s+user_id=\$1\s+AND\s+sku_id=\$2`).
			WithArgs(uint64(1), uint64(20000+i)).
			WillReturnRows(pgxmock.NewRows([]string{"dummy"}).AddRow(1))

		store.DeleteItem(ctx, 1, uint64(20000+i))
	}
	mockPool.ExpectationsWereMet()
}

func BenchmarkStore_ClearCart_QueryStyle(b *testing.B) {
	ctx := context.Background()
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	store := pgstore.New(mockPool)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockPool.ExpectQuery(`(?i)^DELETE\s+FROM\s+cart\s+WHERE\s+user_id=\$1`).
			WithArgs(uint64(1)).
			WillReturnRows(pgxmock.NewRows([]string{"dummy"}).AddRow(1))

		store.ClearCart(ctx, 1)
	}
	mockPool.ExpectationsWereMet()
}

func BenchmarkStore_GetCart(b *testing.B) {
	ctx := context.Background()
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	store := pgstore.New(mockPool)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := pgxmock.NewRows([]string{"sku_id", "count"}).
			AddRow(uint64(1001), uint64(2)).
			AddRow(uint64(1002), uint64(1))

		mockPool.ExpectQuery(`(?i)^SELECT\s+sku_id,\s*count\s+FROM\s+cart\s+WHERE\s+user_id=\$1\s+ORDER\s+BY\s+sku_id`).
			WithArgs(uint64(1)).
			WillReturnRows(rows)

		store.GetCart(ctx, 1)
	}
	mockPool.ExpectationsWereMet()
}
