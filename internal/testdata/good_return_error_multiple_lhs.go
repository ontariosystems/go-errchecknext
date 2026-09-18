package testdata

func goodReturnErrMultipleLHS() error {
	_, err := exampleMulti()
	return err
}
