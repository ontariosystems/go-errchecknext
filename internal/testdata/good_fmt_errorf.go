package testdata

import "fmt"

func goodFmtErrorf() error {
	err := fmt.Errorf("some error")
	sink(err.Error())
	return err
}
