package example

import (
	"fmt"
)

func badSurpressedFollowing() {
	err := example()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = example() //nolint
}
