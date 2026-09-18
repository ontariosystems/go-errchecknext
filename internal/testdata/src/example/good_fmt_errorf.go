package example

import "fmt"

func goodFmtErrorf() error {
	err := fmt.Errorf("some error")
	return err
}
