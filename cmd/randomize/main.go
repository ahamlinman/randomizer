package main // import "go.alexhamlin.co/randomizer/cmd/randomize"

import (
	"fmt"
	"os"

	bolt "go.etcd.io/bbolt"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	boltstore "go.alexhamlin.co/randomizer/pkg/store/bbolt"
)

func main() {
	db, err := bolt.Open("randomizer.db", os.ModePerm&0644, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}

	app := randomizer.NewApp("randomize", boltstore.New(db))
	result, err := app.Main(os.Args[1:])
	if err != nil {
		err := err.(randomizer.Error)
		fmt.Fprintln(os.Stderr, err.HelpText())
		fmt.Fprintf(os.Stderr, "\n%+v\n", err.Cause())
		os.Exit(1)
	}

	fmt.Println(result.Message())
}
