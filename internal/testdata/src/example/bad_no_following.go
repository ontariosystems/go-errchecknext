package example

import "fmt"

func badNoFollowing() {
	err := example()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = example() // want "assignment to err is not immediately followed by an error check"
}
