package example

import (
	"fmt"
)

func badDeclared() error {
	err := example()
	fmt.Println(err.Error()) // want "statement between assignment to err and error check"

	return nil
}
