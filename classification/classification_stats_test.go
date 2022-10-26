package classification_test

import (
	"fmt"
	"log"
	"path"
	"sort"
	"testing"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/classification"
	"github.com/nordicsense/landsat/dataset"
	"github.com/nordicsense/landsat/io"
)

const (
	max  = 25000000
	root = "/Volumes/Caffeine/Data/Landsat/results/v11/classification"
)

func TestCollectClassificationStats(t *testing.T) {
	var (
		err        error
		r          dataset.UniBandReader
		inputTiffs []string
	)

	if inputTiffs, err = io.ScanTree(root, ".*.tiff"); err != nil {
		log.Fatal(err)
	}
	sort.Strings(inputTiffs)
	for _, inputTiff := range inputTiffs {

		if r, err = dataset.OpenUniBand(inputTiff); err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		nx := r.ImageParams().XSize()
		ny := r.ImageParams().YSize()
		row := make([]int8, nx)
		data := make(map[int]int)
		// ugly performance workaround
		ds := r.BreakGlass()
		total := 0
		for y := 0; y < ny; y++ {
			if err = ds.RasterBand(1).IO(gdal.Read, 0, y, nx, 1, row, nx, 1, 0, 0); err != nil {
				t.Fatal(err)
			}
			for x := 0; x < nx; x++ {
				v := row[x]
				if v == 0 {
					continue
				}
				data[int(v)]++
				total++
			}
		}
		fmt.Println("class,id,count,overtotal,overmax")
		fmt.Printf("%s,,%d,1.0,%.3f\n", path.Base(inputTiff), total, float64(total)/float64(max))
		for i := 1; i <= classification.NClasses; i++ {
			fmt.Printf("%s,%d,%d,%.3f,%.3f\n", classification.ClassIdToName[i-1], i, data[i], float64(data[i])/float64(total), float64(data[i])/float64(max))
		}
	}
}
