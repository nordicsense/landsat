package filter

import (
	"os"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"github.com/vardius/progress-go"
)

func Filter5x5(inputTiff, outputTiff string, skip, verbose bool) error {
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

	w, err := dataset.NewUniBand(outputTiff, dataset.GTiff, ip, rp)
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

	for y := 2; y < ny-2; y++ {
		if verbose {
			bar.Advance(1)
		}
		row := make([]int8, nx)
		buf := make([]int8, 25)
		for x := 2; x < nx-2; x++ {
			if err = ds.IO(gdal.Read, x-2, y-2, 5, 5, buf, 5, 5, 0, 0); err != nil {
				return err
			}
			var (
				val   int8
				count int
			)
			freq := make(map[int8]int)
			for i := 0; i < 25; i++ {
				if i == 0 || i == 4 || i == 20 || i == 24 {
					continue
				}
				pixval := buf[i]
				freq[pixval]++
				if i == 12 {
					freq[pixval] += 4 // weigh 3x for the middle point
				} else if i == 11 || i == 13 || i == 7 || i == 17 {
					freq[pixval] += 2 // weigh 1up/down and 1left/right
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
