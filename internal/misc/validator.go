package misc

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var varNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func VarName(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return varNameRegex.MatchString(field.String())
	default:
		return false
	}
}

var _validator = validator.New()

func init() {
	if err := _validator.RegisterValidation("varname", VarName); err != nil {
		panic(err)
	}
}

func Validator() *validator.Validate {
  return _validator
}
