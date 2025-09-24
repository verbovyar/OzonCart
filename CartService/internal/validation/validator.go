package validation

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	v := validator.New()

	v.RegisterValidation("id", func(fl validator.FieldLevel) bool {
		switch fl.Field().Kind() {
		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fl.Field().Int() > 0
		default:
			return false
		}
	})

	alnumSpace := regexp.MustCompile(`^[\p{L}\p{N}\s\-\_]*$`)
	v.RegisterValidation("alnumspace", func(fl validator.FieldLevel) bool {
		s, ok := fl.Field().Interface().(string)
		return ok && alnumSpace.MatchString(s)
	})

	return &Validator{v: v}
}

func (v *Validator) Struct(s any) error {
	return v.v.Struct(s)
}

func (v *Validator) ValidateID(id uint64) bool {
	return id > 0
}
