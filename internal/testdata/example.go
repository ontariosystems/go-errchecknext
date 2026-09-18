package testdata

import (
	"io"
)

func example() error {
	return nil
}

func exampleMulti() (int, error) { return 0, nil }

type testIO struct{}

func (*testIO) Close() error { return nil }
func exampleClosable() (io.Closer, error) {
	return &testIO{}, nil
}
