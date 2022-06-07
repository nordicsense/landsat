package stats

import (
	"fmt"
	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"github.com/nordicsense/landsat/training"
	"path"
)

func Collect(inputTiff string, max int) error {
	var (
		err error
		r   dataset.UniBandReader
	)

	if r, err = dataset.OpenUniBand(inputTiff); err != nil {
		return err
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
			return err
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
	for i := 1; i <= training.NClasses; i++ {
		fmt.Printf("%s,%d,%d,%.3f,%.3f\n", training.ClassIdToName[i-1], i, data[i], float64(data[i])/float64(total), float64(data[i])/float64(max))
	}
	return nil
}
