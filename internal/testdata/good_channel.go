package testdata

func goodChannel() error {
	errChan := make(chan error)

	err := example()
	errChan <- err

	return <-errChan
}
