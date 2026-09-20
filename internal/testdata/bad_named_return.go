package testdata

var errBadBareReturn error

func badBareReturn() (err error) {
	errBadBareReturn = example()
	return // want "statement between assignment to err and error check"
}
