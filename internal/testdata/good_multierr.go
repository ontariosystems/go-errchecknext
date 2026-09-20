package testdata

import "github.com/hashicorp/go-multierror"

func goodMultierr() error {
	var errs error
	if err := example(); err != nil {
		errs = multierror.Append(errs, err)
		_ = errs
	}
	return errs
}
