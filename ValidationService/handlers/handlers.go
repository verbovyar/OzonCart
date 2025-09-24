package handlers

import (
	"context"
	"validation/api/ValidationServiceApiPb"
	"validation/infrastructure/cartServiceClient/api/CartServiceApiPb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ValidationRouter struct {
	ValidationServiceApiPb.UnimplementedValidationServiceServer
	cartClient CartServiceApiPb.CartServiceClient
}

func NewValidationRouter(cartClient CartServiceApiPb.CartServiceClient) *ValidationRouter {
	return &ValidationRouter{cartClient: cartClient}
}

func forwardMeta(ctx context.Context, hdr, trl metadata.MD) {
	if len(hdr) > 0 {
		_ = grpc.SetHeader(ctx, hdr)
	}
	if len(trl) > 0 {
		grpc.SetTrailer(ctx, trl)
	}
}

func (v *ValidationRouter) AddToCart(ctx context.Context, in *ValidationServiceApiPb.AddToCartRequest) (*ValidationServiceApiPb.GetCartResponse, error) {
	if in.GetUserId() == 0 || in.GetSkuId() == 0 || in.GetCount() == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	var hdr, trl metadata.MD

	res, err := v.cartClient.AddToCart(ctx, &CartServiceApiPb.AddToCartRequest{
		UserId: in.UserId,
		SkuId:  in.SkuId,
		Count:  in.Count,
	}, grpc.Header(&hdr), grpc.Trailer(&trl))

	if err != nil {
		return nil, err
	}

	forwardMeta(ctx, hdr, trl)

	return toValResp(res), nil
}

func (v *ValidationRouter) DeleteItem(ctx context.Context, in *ValidationServiceApiPb.DeleteItemRequest) (*ValidationServiceApiPb.GetCartResponse, error) {
	if in.GetUserId() == 0 || in.GetSkuId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	var hdr, trl metadata.MD

	res, err := v.cartClient.DeleteItem(ctx, &CartServiceApiPb.DeleteItemRequest{
		UserId: in.UserId,
		SkuId:  in.SkuId,
	}, grpc.Header(&hdr), grpc.Trailer(&trl))

	if err != nil {
		return nil, err
	}

	forwardMeta(ctx, hdr, trl)

	return toValResp(res), nil
}

func (v *ValidationRouter) ClearCart(ctx context.Context, in *ValidationServiceApiPb.ClearCartRequest) (*ValidationServiceApiPb.GetCartResponse, error) {
	if in.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	var hdr, trl metadata.MD

	res, err := v.cartClient.ClearCart(ctx, &CartServiceApiPb.ClearCartRequest{UserId: in.UserId}, grpc.Trailer(&trl))

	if err != nil {
		return nil, err
	}

	forwardMeta(ctx, hdr, trl)

	return toValResp(res), nil
}

func (v *ValidationRouter) GetCart(ctx context.Context, in *ValidationServiceApiPb.GetCartRequest) (*ValidationServiceApiPb.GetCartResponse, error) {
	if in.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	var hdr, trl metadata.MD

	res, err := v.cartClient.GetCart(ctx, &CartServiceApiPb.GetCartRequest{UserId: in.UserId}, grpc.Trailer(&trl))
	if err != nil {
		return nil, err
	}

	forwardMeta(ctx, hdr, trl)

	return toValResp(res), nil
}

func toValResp(c *CartServiceApiPb.GetCartResponse) *ValidationServiceApiPb.GetCartResponse {
	out := &ValidationServiceApiPb.GetCartResponse{
		TotalPrice: c.TotalPrice,
	}

	for _, it := range c.Items {
		out.Items = append(out.Items, &ValidationServiceApiPb.CartItem{
			SkuId: it.SkuId,
			Name:  it.Name,
			Count: it.Count,
			Price: it.Price,
		})
	}

	return out
}
