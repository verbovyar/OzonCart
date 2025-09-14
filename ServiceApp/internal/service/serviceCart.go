package service

import (
	"context"
	"errors"

	"github.com/verbovyar/OzonCart/internal/domain"
	"github.com/verbovyar/OzonCart/internal/repositories/interfaces"
)

var ErrProductNotFound = errors.New("product not found")

type CartService struct {
	store interfaces.RepositoryIface
	pc    ClientIface
}

func New(store interfaces.RepositoryIface, pc ClientIface) *CartService {
	return &CartService{
		store: store,
		pc:    pc,
	}
}

func (c *CartService) AddToCart(ctx context.Context, userID, skuID, count uint64) error {
	_, err := c.pc.GetProduct(ctx, skuID)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	return c.store.AddItem(ctx, userID, skuID, count)
}

func (c *CartService) DeleteItem(userID, skuID uint64) {
	c.store.DeleteItem(context.Background(), userID, skuID)
}

func (c *CartService) ClearCart(userID uint64) {
	c.store.ClearCart(context.Background(), userID)
}

func (c *CartService) GetCart(ctx context.Context, userID uint64) (*domain.GetCartResponse, error) {
	positions, err := c.store.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.CartItem, 0, len(positions))
	var total uint64
	for _, p := range positions {
		pr, err := c.pc.GetProduct(ctx, p.SkuID)
		if err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return nil, err
		}
		item := domain.CartItem{
			SkuID: p.SkuID,
			Name:  pr.Name,
			Count: p.Count,
			Price: pr.Price,
		}
		items = append(items, item)
		total += pr.Price * p.Count
	}

	return &domain.GetCartResponse{Items: items, TotalPrice: total}, nil
}
