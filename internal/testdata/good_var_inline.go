package testdata

func goodVarInline() error {
	var err error

	if err = example(); err != nil {
		return err
	}

	return nil
}
