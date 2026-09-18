package testdata

import (
	"github.com/hashicorp/errwrap"
)

func goodErrWrap() error {
	err := errwrap.Wrap(example(), example())
	_ = err
	return err
}

func goodErrWrapf() error {
	err := errwrap.Wrapf("wrapped {{err}}", example())
	_ = err
	return err
}
