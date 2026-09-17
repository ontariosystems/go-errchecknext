package example

import "fmt"

func badSurpressedLinterFollowing() {
	err := example()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = example() //nolint:errchecknext
}
