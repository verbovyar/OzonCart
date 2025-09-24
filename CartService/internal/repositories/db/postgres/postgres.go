package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgxPoolIface interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Position struct {
	SkuID uint64
	Count uint64
}

type Store struct {
	pool PgxPoolIface
}

func New(pool PgxPoolIface) *Store {
	return &Store{pool: pool}
}

func (s *Store) AddItem(ctx context.Context, userID, skuID, count uint64) error {
	query := `INSERT INTO Cart (user_id, sku_id, count) VALUES ($1, $2, $3) 
				ON CONFLICT (user_id, sku_id) DO UPDATE SET count = cart.count + EXCLUDED.count`
	rows, err := s.pool.Query(ctx, query, userID, skuID, count)
	if err != nil {
		return err
	}
	defer rows.Close()

	return err
}

func (s *Store) DeleteItem(ctx context.Context, userID, skuID uint64) error {
	query := `DELETE FROM Cart WHERE user_id=$1 AND sku_id=$2`
	rows, err := s.pool.Query(ctx, query, userID, skuID)
	if err != nil {
		return err
	}
	defer rows.Close()

	return err
}

func (s *Store) ClearCart(ctx context.Context, userID uint64) error {
	query := `DELETE FROM Cart WHERE user_id=$1`
	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	return err
}

func (s *Store) GetCart(ctx context.Context, userID uint64) ([]Position, error) {
	query := `SELECT sku_id, count FROM Cart WHERE user_id=$1 ORDER BY sku_id`
	rows, err := s.pool.Query(ctx, query, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ans []Position
	for rows.Next() {
		var temp Position
		if err := rows.Scan(&temp.SkuID, &temp.Count); err != nil {
			return nil, err
		}
		ans = append(ans, temp)
	}

	return ans, nil
}
