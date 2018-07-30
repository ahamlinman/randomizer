package main

import (
	"fmt"
	"os"

	"github.com/ahamlinman/randomizer"
)

func main() {
	result, err := randomizer.Main(os.Args[1:])

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Println(result)
}
