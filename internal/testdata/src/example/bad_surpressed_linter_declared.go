package example

import "fmt"

func badSurpressedLinterDeclared() error {
	err := example()
	fmt.Println(err.Error()) //nolint:errchecknext

	return nil
}
