package handlers

import (
	"context"
	"errors"
	"sync"

	"github.com/verbovyar/OzonCart/api/CartServiceApiPb"
	"github.com/verbovyar/OzonCart/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartGrpcRouter struct {
	cs *service.CartService

	wg sync.Mutex

	CartServiceApiPb.UnimplementedCartServiceServer
}

func NewGrpsRouter(cs *service.CartService) *CartGrpcRouter {
	return &CartGrpcRouter{cs: cs}
}

func (c *CartGrpcRouter) AddToCart(ctx context.Context, in *CartServiceApiPb.AddToCartRequest) (*CartServiceApiPb.GetCartResponse, error) {
	c.wg.Lock()
	defer c.wg.Unlock()

	err := c.cs.AddToCart(ctx, in.UserId, in.SkuId, in.Count)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, "db error")
	}

	return &CartServiceApiPb.GetCartResponse{}, nil
}
func (c *CartGrpcRouter) DeleteItem(ctx context.Context, in *CartServiceApiPb.DeleteItemRequest) (*CartServiceApiPb.GetCartResponse, error) {
	c.wg.Lock()
	defer c.wg.Unlock()

	err := c.cs.DeleteItem(in.UserId, in.SkuId)

	if errors.Is(err, service.ErrProductNotFound) {
		return nil, status.Error(codes.NotFound, "item or cart not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &CartServiceApiPb.GetCartResponse{}, nil
}

func (c *CartGrpcRouter) ClearCart(ctx context.Context, in *CartServiceApiPb.ClearCartRequest) (*CartServiceApiPb.GetCartResponse, error) {
	c.wg.Lock()
	defer c.wg.Unlock()

	err := c.cs.ClearCart(in.UserId)

	if errors.Is(err, service.ErrProductNotFound) {
		return nil, status.Error(codes.NotFound, "cart not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &CartServiceApiPb.GetCartResponse{}, nil
}

func (c *CartGrpcRouter) GetCart(ctx context.Context, in *CartServiceApiPb.GetCartRequest) (*CartServiceApiPb.GetCartResponse, error) {
	c.wg.Lock()
	defer c.wg.Unlock()

	cart, err := c.cs.GetCart(ctx, in.UserId)

	if err != nil {
		return nil, status.Error(codes.Internal, "db error")
	}

	items := make([]*CartServiceApiPb.CartItem, 0)
	for _, temp_item := range cart.Items {
		items = append(items, &CartServiceApiPb.CartItem{
			SkuId: temp_item.SkuID,
			Name:  temp_item.Name,
			Count: temp_item.Count,
			Price: temp_item.Price,
		})
	}

	return &CartServiceApiPb.GetCartResponse{
		Items:      items,
		TotalPrice: cart.TotalPrice,
	}, nil
}
