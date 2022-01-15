package main

import (
	"github.com/nordicsense/landsat"
	"log"
	"os"
	"strconv"
)

func main() {
	clip, _ := strconv.ParseBool(os.Args[3])

	// run with e.g. compress=deflate zlevel=6 predictor=3
	// best for float32, see https://kokoalberti.com/articles/geotiff-compression-optimization-guide/
	options := os.Args[4:]
	if err := landsat.MergeAndCorrect(os.Args[1], os.Args[2], clip, options...); err != nil {
		log.Fatal(err)
	}
}
