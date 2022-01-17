package main

import (
	"log"
	"os"
)

func main() {
	options := os.Args[3:]
	if err := HistCorrect(os.Args[1], os.Args[2], options...); err != nil {
		log.Fatal(err)
	}
	/*
		if err := HistCollect(os.Args[1], os.Args[2]); err != nil {
			log.Fatal(err)
		}
		if err := HistCollect(os.Args[1], os.Args[2]+"_histcorr"); err != nil {
			log.Fatal(err)
		}
	*/
}
