package testdata

func badSurpressedFollowing() {
	err := example()
	if err != nil {
		sink(err.Error())
	}

	err = example() //nolint
}
