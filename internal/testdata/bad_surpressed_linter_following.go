package testdata

func badSurpressedLinterFollowing() {
	err := example()
	if err != nil {
		sink(err.Error())
	}

	err = example() //nolint:errchecknext
}
