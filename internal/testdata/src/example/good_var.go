package example

func goodVar() error {
	var err error
	err = example()

	if err != nil {
		return err
	}

	return nil
}
