package testdata

func badDeclared() error {
	err := example()
	sink(err) // want "statement between assignment to err and error check"

	return nil
}
