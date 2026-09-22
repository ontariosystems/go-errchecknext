package testdata

func badChannel() error {
	errChan := make(chan error, 1)

	err := example()
	sink(err) // want "statement between assignment to err and error check"
	errChan <- err

	return <-errChan
}
