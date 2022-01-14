package main

import (
	"github.com/nordicsense/landsat"
	"log"
	"os"
)

func main() {
	if err := landsat.MergeAndCorrect(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
