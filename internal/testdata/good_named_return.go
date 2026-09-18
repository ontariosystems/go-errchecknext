package testdata

func goodNamedReturn() (err error) {
	if err = example(); err != nil {
		return
	}
	return nil
}
