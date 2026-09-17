package example

import "io"

func example() error {
	return nil
}

type testIO struct{}

func (*testIO) Close() error { return nil }
func exampleClosable() (io.Closer, error) {
	return &testIO{}, nil
}
