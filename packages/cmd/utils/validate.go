package utils

import "github.com/go-playground/validator/v10"

var Validate = validator.New(validator.WithRequiredStructEnabled())

func SelectFirstString(vals ...string) string {
	for _, val := range vals {
		if val != "" {
			return val
		}
	}

	return vals[len(vals)-1]
}
