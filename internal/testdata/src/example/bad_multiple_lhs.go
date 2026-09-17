package example

func badMultipleLHS() error {
	c, err := exampleClosable()
	defer func() { _ = c.Close() }() // want "statement between assignment to err and error check"
	if err != nil {
		return err
	}

	return nil
}
