package service_test

import (
	"context"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"github.com/verbovyar/OzonCart/internal/mocks"
	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
	"github.com/verbovyar/OzonCart/internal/service"
)

func TestCartService_Add_Succes(t *testing.T) {
	mc := minimock.NewController(t)

	ctx := context.Background()
	userID := uint64(1)
	skuID := uint64(1001)
	count := uint64(2)

	pc := mocks.NewClientIfaceMock(mc)
	repo := mocks.NewRepositoryIfaceMock(mc)

	pc.GetProductMock.Expect(ctx, skuID).Return(&service.Product{Name: "Demo T-Shirt", Price: 1500}, nil)
	repo.AddItemMock.Expect(ctx, userID, skuID, uint64(count)).Return(nil)

	cs := service.New(repo, pc)
	require.NoError(t, cs.AddToCart(ctx, userID, skuID, count))
}

func TestCartService_Add_Not_Found(t *testing.T) {
	mc := minimock.NewController(t)

	ctx := context.Background()
	userID := uint64(1)
	skuID := uint64(21)
	count := uint64(2)

	pc := mocks.NewClientIfaceMock(mc)
	repo := mocks.NewRepositoryIfaceMock(mc)

	pc.GetProductMock.Expect(ctx, skuID).Return(nil, service.ErrProductNotFound)

	cs := service.New(repo, pc)
	err := cs.AddToCart(ctx, userID, skuID, count)
	require.ErrorIs(t, err, service.ErrProductNotFound)
}

func TestCartService_GetCart_AllProducts(t *testing.T) {
	mc := minimock.NewController(t)

	userID := uint64(7)

	repo := mocks.NewRepositoryIfaceMock(mc)
	pc := mocks.NewClientIfaceMock(mc)

	repo.GetCartMock.Expect(minimock.AnyContext, userID).Return(
		[]postgres.Position{
			{SkuID: 1001, Count: 2},
			{SkuID: 1002, Count: 1},
		}, nil,
	)

	pc.GetProductMock.When(minimock.AnyContext, uint64(1001)).
		Then(&service.Product{Name: "Demo T-Shirt", Price: 1500}, nil)

	pc.GetProductMock.When(minimock.AnyContext, uint64(1002)).
		Then(&service.Product{Name: "Coffee Mug", Price: 900}, nil)

	cs := service.New(repo, pc)
	res, err := cs.GetCart(context.Background(), userID)
	t.Logf("%d", res.TotalPrice)

	require.NoError(t, err)
	require.Len(t, res.Items, 2)
	require.Equal(t, uint64(3900), res.TotalPrice)
}

func TestCartService_GetCart_SkipMissing(t *testing.T) {
	mc := minimock.NewController(t)

	userID := uint64(7)

	repo := mocks.NewRepositoryIfaceMock(mc)
	pc := mocks.NewClientIfaceMock(mc)

	repo.GetCartMock.Expect(minimock.AnyContext, userID).Return(
		[]postgres.Position{
			{SkuID: 1001, Count: 2},
			{SkuID: 1002, Count: 1},
		}, nil,
	)

	pc.GetProductMock.When(minimock.AnyContext, uint64(1001)).
		Then(&service.Product{Name: "Demo T-Shirt", Price: 1500}, nil)

	pc.GetProductMock.When(minimock.AnyContext, uint64(1002)).
		Then(nil, service.ErrProductNotFound)

	cs := service.New(repo, pc)
	res, err := cs.GetCart(context.Background(), userID)

	require.NoError(t, err)
	require.Len(t, res.Items, 1)
	require.Equal(t, uint64(3000), res.TotalPrice)
}
