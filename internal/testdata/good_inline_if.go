package testdata

func goodInlineIf() error {
	if err := example(); err != nil {
		return err
	}

	return nil
}
