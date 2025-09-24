package interfaces

import (
	"context"

	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
)

// internal/repositories/interfaces/interfaces.go
//
//go:generate minimock -i RepositoryIface -o ../../mocks   -s "_mock.go"
type RepositoryIface interface {
	AddItem(ctx context.Context, userID, skuID uint64, count uint64) error
	DeleteItem(ctx context.Context, userID, skuID uint64) error
	ClearCart(ctx context.Context, userID uint64) error
	GetCart(ctx context.Context, userID uint64) ([]postgres.Position, error)
}
