package docs

import "github.com/swaggo/swag"

func init() {
	swag.Register("Cart service swagger", &s)
}

var s swag.Spec
