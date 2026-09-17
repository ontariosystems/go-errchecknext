package example

func goodDeclared() error {
	err := example()

	if err != nil {
		return err
	}

	return nil
}
