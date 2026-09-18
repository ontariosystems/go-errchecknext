package testdata

import (
	"errors"
)

func goodErrorsJoin() error {
	var errs error
	if err := example(); err != nil {
		errs = errors.Join(errs, err)
		_ = errs
	}
	return errs
}
