package testdata

func badVar() error {
	var err error
	err = example()

	sink(err.Error()) // want "statement between assignment to err and error check"

	return nil
}
