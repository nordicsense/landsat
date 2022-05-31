package trim

import (
	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"github.com/vardius/progress-go"
	"os"
)

var (
	TL = dataset.LatLon{7668849, 479358} // sin
	TR = dataset.LatLon{7620132, 592293}
	BR = dataset.LatLon{7341672, 464008}
	BL = dataset.LatLon{7392573, 358017}
)

func Process(inputTiff, outputTiff string, skip, verbose bool, tl, tr, br, bl dataset.LatLon) error {
	if _, err := os.Stat(outputTiff); skip && err == nil {
		return nil
	}

	var (
		err error
		r   dataset.UniBandReader
	)

	if r, err = dataset.OpenUniBand(inputTiff); err != nil {
		return err
	}
	defer r.Close()

	ip := r.ImageParams()
	isAboveTop := leftOf(tl, tr, ip)
	isBelowBottom := leftOf(br, bl, ip)
	isLeftOfLeft := leftOf(bl, tl, ip)
	isRightOfRight := leftOf(tr, br, ip)

	// ugly performance workaround
	ds := r.BreakGlass()

	w, err := dataset.NewUniBand(outputTiff, dataset.GTiff,
		ip.ToBuilder().DataType(gdal.Byte).NaN(0.).Build(),
		r.RasterParams().ToBuilder().Offset(0.).Scale(1.).Build(),
		"compress=LZW", "predictor=2")
	if err != nil {
		return err
	}
	defer w.Close()

	wds := w.BreakGlass().RasterBand(1)

	nx := ip.XSize()
	ny := ip.YSize()

	bar := progress.New(0, int64(ny))
	if verbose {
		bar.Start()
	}

	row := make([]int8, nx)
	for y := 0; y < ny; y++ {
		if err = ds.RasterBand(1).IO(gdal.Read, 0, y, nx, 1, row, nx, 1, 0, 0); err != nil {
			return err
		}
		for x := 0; x < nx; x++ {
			if row[x] > 0 {
				if isAboveTop(x, y) || isBelowBottom(x, y) || isLeftOfLeft(x, y) || isRightOfRight(x, y) {
					row[x] = 0
				}
			}
		}
		if err = wds.IO(gdal.Write, 0, y, nx, 1, row, nx, 1, 0, 0); err != nil {
			return err
		}
		if verbose {
			bar.Advance(1)
		}
	}
	if verbose {
		bar.Stop()
	}
	return err
}

func leftOf(ll1, ll2 dataset.LatLon, ip *dataset.ImageParams) func(x, y int) bool {
	x1, y1 := ip.Transform().LatLonSin2Pixels(ll1)
	x2, y2 := ip.Transform().LatLonSin2Pixels(ll2)
	y2y1 := float64(y2 - y1)
	x2x1 := float64(x2 - x1)
	return func(x, y int) bool {
		v := float64(x-x1)*y2y1 - float64(y-y1)*x2x1
		return v >= 0.0
	}
}
