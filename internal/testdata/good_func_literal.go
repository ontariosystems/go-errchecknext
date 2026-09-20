package testdata
func goodFuncLiteral() {
	fn := func() error {
		err := example()
		if err != nil {
			return err
		}
		return nil
	}

	_ = fn()
}