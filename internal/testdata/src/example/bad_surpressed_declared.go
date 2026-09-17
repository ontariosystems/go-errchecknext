package example

import "fmt"

func badSurpressedDeclared() error {
	err := example()
	fmt.Println(err.Error()) //nolint

	return nil
}
