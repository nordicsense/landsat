package main

import (
	"log"
	"os"
)

func main() {
	// run with e.g. compress=deflate zlevel=6 predictor=3
	// best for float32, see https://kokoalberti.com/articles/geotiff-compression-optimization-guide/
	var options []string
	if len(os.Args) > 3 {
		options = os.Args[3:]
	}
	if err := MergeAndCorrect(os.Args[1], os.Args[2], options...); err != nil {
		log.Fatal(err)
	}
}
