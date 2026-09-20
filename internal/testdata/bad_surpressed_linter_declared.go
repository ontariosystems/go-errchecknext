package testdata

func badSurpressedLinterDeclared() error {
	err := example()
	sink(err.Error()) //nolint:errchecknext

	return nil
}
