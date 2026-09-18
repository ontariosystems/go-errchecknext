package testdata

import "fmt"

func goodReturnErrorWrapped() error {
	err := fmt.Errorf("wrapped %w", example())
	return err
}
