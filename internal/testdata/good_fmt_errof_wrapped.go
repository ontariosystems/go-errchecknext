package testdata

import "fmt"

func goodFmtErrorfWrapped() error {
	err := fmt.Errorf("wrapped %w", example())
	return err
}
