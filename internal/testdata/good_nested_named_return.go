package testdata

func goodNestedNamedReturn() (err error) {
	{
		err = example()
		return
	}
}
