package example

func badNamedReturn() (err error) {
	err = example()
	return nil // want "statement between assignment to err and error check"
}
