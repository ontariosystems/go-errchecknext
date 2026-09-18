package example

func goodIgnoreErrorMultipleLHS() error {
	c, _ := exampleClosable()
	defer func() { _ = c.Close() }()
	return nil
}
