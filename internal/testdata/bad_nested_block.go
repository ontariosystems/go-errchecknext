package testdata

func badNestedBlock() error {
	{
		errBadNested := example()
		sink(errBadNested) // want "statement between assignment to err and error check"
	}
	return nil
}
