package testdata

var errNestedNamedReturn error

func badNestedNamedReturn() (err error) {
	{
		errNestedNamedReturn = example()
		sink(errNestedNamedReturn) // want "statement between assignment to err and error check"
	}
	return
}
