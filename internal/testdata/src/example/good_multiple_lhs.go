package example

func goodMultipleLHS() error {
	c, err := exampleClosable()
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	return nil
}
