package testdata

import "fmt"

func goodFmtErrorf() error {
	err := fmt.Errorf("some error")
	fmt.Println(err.Error())
	return err
}
