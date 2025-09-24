package handlers

import (
	"context"
	"errors"

	"github.com/verbovyar/OzonCart/api/CartServiceApiPb"
	"github.com/verbovyar/OzonCart/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CartGrpcRouter struct {
	cs *service.CartService

	CartServiceApiPb.UnimplementedCartServiceServer
}

func NewGrpsRouter(cs *service.CartService) *CartGrpcRouter {
	return &CartGrpcRouter{cs: cs}
}

func (c *CartGrpcRouter) addToCart(ctx context.Context, in *CartServiceApiPb.AddToCartRequest) (*emptypb.Empty, error) {
	err := c.cs.AddToCart(ctx, in.UserId, in.SkuId, in.Count)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, "db error")
	}

	return &emptypb.Empty{}, status.Error(codes.OK, "created")
}

func (c *CartGrpcRouter) deleteItem(ctx context.Context, in *CartServiceApiPb.DeleteItemRequest) (*emptypb.Empty, error) {
	err := c.cs.DeleteItem(in.UserId, in.SkuId)

	if errors.Is(err, service.ErrProductNotFound) {
		return nil, status.Error(codes.NotFound, "item or cart not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (c *CartGrpcRouter) clearCart(ctx context.Context, in *CartServiceApiPb.ClearCartRequest) (*emptypb.Empty, error) {
	err := c.cs.ClearCart(in.UserId)

	if errors.Is(err, service.ErrProductNotFound) {
		return nil, status.Error(codes.NotFound, "cart not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (c *CartGrpcRouter) getCart(ctx context.Context, in *CartServiceApiPb.GetCartRequest) (*CartServiceApiPb.GetCartResponse, error) {
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
