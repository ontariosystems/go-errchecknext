package testdata

func badNoFollowing() {
	err := example()
	if err != nil {
		sink(err.Error())
	}

	err = example() // want "assignment to err is not immediately followed by an error check"
}
