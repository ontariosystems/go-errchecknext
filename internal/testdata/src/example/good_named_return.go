package example

func goodNamedReturn() (err error) {
	if err = example(); err != nil {
		return
	}
	return nil
}
