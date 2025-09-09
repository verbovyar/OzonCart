package domain

type AddToCartRequest struct {
	Count uint64 `json:"count" validate:"required,gt=0,lte=60000"`
}

type CartItem struct {
	SkuID uint64 `json:"sku_id"`
	Name  string `json:"name"`
	Count uint64 `json:"count"`
	Price uint64 `json:"price"`
}

type GetCartResponse struct {
	Items      []CartItem `json:"items"`
	TotalPrice uint64     `json:"total_price"`
}
