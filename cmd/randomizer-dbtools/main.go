/*

The randomizer-dbtools command provides quick-and-dirty helper utilities for
randomizer data stores.

*/
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "randomizer-dbtools",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
