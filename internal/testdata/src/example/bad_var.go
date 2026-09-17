package example

import "fmt"

func badVar() error {
	var err error
	err = example()

	fmt.Println(err.Error()) // want "statement between assignment to err and error check"

	return nil
}
