package conversion_test

import (
	"github.com/nordicsense/landsat/dataset"
	"github.com/nordicsense/landsat/io"
	"log"
	"math"
	"path"
	"strings"
	"testing"
)

const (
	root    = "/Volumes/Caffeine/Data/Landsat/converted/prod"
	xOffset = 2000
	yOffset = 2000
	dx      = 4000
	dy      = 4000
)

func TestCollectHistograms(t *testing.T) {
	var (
		err        error
		r          dataset.MultiBandReader
		inputTiffs []string
	)

	if inputTiffs, err = io.ScanTree(root, ".*.tiff"); err != nil {
		log.Fatal(err)
	}

	buckets := make(map[int]map[int][100]float64)
	buckets[5] = make(map[int][100]float64)
	buckets[7] = make(map[int][100]float64)
	buckets[8] = make(map[int][100]float64)
	counts := make(map[int]map[int]int)
	counts[5] = make(map[int]int)
	counts[7] = make(map[int]int)
	counts[8] = make(map[int]int)

	for _, inputTiff := range inputTiffs {
		if r, err = dataset.OpenMultiBand(inputTiff); err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		id := 5 // LT05
		if strings.HasPrefix(path.Base(inputTiff), "LE07") {
			id = 7
		} else if strings.HasPrefix(path.Base(inputTiff), "LC08") {
			id = 8
		}

		for band := 1; band <= 7; band++ {
			if id == 8 && band == 1 {
				continue
			}
			if (id == 5 || id == 7) && band == 6 {
				continue
			}
			log.Println(path.Base(inputTiff), id, band)
			var buf []float64
			buf, err = r.Reader(band).ReadBlock(xOffset, yOffset, dataset.Box{0, 0, dx, dy})

			for _, v := range buf {
				if math.IsNaN(v) {
					continue
				}
				counts[id][band]++
				ind := int(v * 100)
				if ind < 0 {
					ind = 0
				} else if ind > 99 {
					ind = 99
				}
				hist := buckets[id][band]
				hist[ind] += 1.
				buckets[id][band] = hist
			}
		}
	}
	for id, sv := range buckets {
		for band, hist := range sv {
			factor := 1. / float64(counts[id][band])
			for j, v := range hist {
				hist[j] = v * factor
			}
			buckets[id][band] = hist
		}
	}
	log.Println("5:1", buckets[5][1])
	log.Println("5:2", buckets[5][2])
	log.Println("5:3", buckets[5][3])
	log.Println("5:4", buckets[5][4])
	log.Println("5:5", buckets[5][5])
	log.Println("5:7", buckets[5][7])

	log.Println("7:1", buckets[7][1])
	log.Println("7:2", buckets[7][2])
	log.Println("7:3", buckets[7][3])
	log.Println("7:4", buckets[7][4])
	log.Println("7:5", buckets[7][5])
	log.Println("7:7", buckets[7][7])

	log.Println("8:2", buckets[8][2])
	log.Println("8:3", buckets[8][3])
	log.Println("8:4", buckets[8][4])
	log.Println("8:5", buckets[8][5])
	log.Println("8:6", buckets[8][6])
	log.Println("8:7", buckets[8][7])
}
