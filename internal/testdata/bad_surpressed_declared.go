package testdata

func badSurpressedDeclared() error {
	err := example()
	sink(err.Error()) //nolint

	return nil
}
