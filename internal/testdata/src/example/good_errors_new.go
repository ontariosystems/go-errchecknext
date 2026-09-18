package example

import "errors"

func goodErrorsNew() error {
	err := errors.New("some error")
	return err
}
