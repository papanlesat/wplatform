package validation

import (
	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate

func Init() {
	Validator = validator.New()
}

func ValidateStruct(s interface{}) error {
	return Validator.Struct(s)
}

func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			field := fieldErr.Field()
			tag := fieldErr.Tag()

			switch tag {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = field + " must be a valid email"
			case "min":
				errors[field] = field + " must be at least " + fieldErr.Param() + " characters"
			case "max":
				errors[field] = field + " must be at most " + fieldErr.Param() + " characters"
			case "oneof":
				errors[field] = field + " must be one of: " + fieldErr.Param()
			default:
				errors[field] = field + " is invalid"
			}
		}
	}

	return errors
}
