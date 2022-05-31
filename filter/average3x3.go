package filter

import (
	"os"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"github.com/vardius/progress-go"
)

func Filter3x3(inputTiff, outputTiff string, skip, verbose bool) error {
	if _, err := os.Stat(outputTiff); skip && err == nil {
		return nil
	}

	r, err := dataset.OpenUniBand(inputTiff)
	if err != nil {
		return err
	}
	defer r.Close()

	// ugly performance workaround
	ds := r.BreakGlass().RasterBand(1)

	ip := r.ImageParams().ToBuilder().Build()
	rp := r.RasterParams().ToBuilder().Build()

	w, err := dataset.NewUniBand(outputTiff, dataset.GTiff, ip, rp, "compress=LZW", "predictor=2")
	if err != nil {
		return err
	}
	defer w.Close()

	wds := w.BreakGlass().RasterBand(1)

	nx := r.ImageParams().XSize()
	ny := r.ImageParams().YSize()

	bar := progress.New(0, int64(ny))
	if verbose {
		bar.Start()
	}

	for y := 1; y < ny-1; y++ {
		if verbose {
			bar.Advance(1)
		}
		row := make([]int8, nx)
		buf := make([]int8, 9)
		for x := 1; x < nx-1; x++ {
			if err = ds.IO(gdal.Read, x-1, y-1, 3, 3, buf, 3, 3, 0, 0); err != nil {
				return err
			}
			var (
				val   int8
				count int
			)
			freq := make(map[int8]int)
			for i := 0; i < 9; i++ {
				pixval := buf[i]
				freq[pixval]++
				if i == 4 {
					freq[pixval] += 2 // weigh 3x for the middle point
				}
				if freq[pixval] > count {
					count = freq[pixval]
					val = pixval
				}
			}
			row[x] = val
		}

		err = wds.IO(gdal.Write, 0, y, nx, 1, row, nx, 1, 0, 0)
		if err != nil {
			return err
		}
	}
	if verbose {
		bar.Stop()
	}
	return nil
}
