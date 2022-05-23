package tensorflow

import (
	"fmt"
	"math"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/data"
	"github.com/nordicsense/landsat/dataset"
	"github.com/vardius/progress-go"
)

func Predict(modelDir, inputTiff, outputTiff string, minx, maxx, miny, maxy int, isL8 bool) error {
	model, err := LoadModel(modelDir)
	if err != nil {
		return err
	}
	defer model.Close()

	r, err := dataset.OpenMultiBand(inputTiff)
	if err != nil {
		return err
	}
	defer r.Close()

	// ugly performance workaround
	ds := r.Reader(1).BreakGlass()

	ip := r.ImageParams().ToBuilder().DataType(gdal.Byte).NaN(0.).Build()
	rp := r.Reader(1).RasterParams().ToBuilder().Offset(0.).Scale(1.).Build()

	w, err := dataset.NewUniBand(outputTiff, dataset.GTiff, ip, rp)
	if err != nil {
		return err
	}
	defer w.Close()

	wds := w.BreakGlass().RasterBand(1)

	nx := r.ImageParams().XSize()
	if minx < 0 {
		minx = 0
	}
	if maxx > nx {
		maxx = nx
	}
	dx := maxx - minx
	ny := r.ImageParams().YSize()
	if miny < 0 {
		miny = 0
	}
	if maxy > ny {
		maxy = ny
	}

	bar := progress.New(int64(miny), int64(maxy))
	bar.Start()

	var rrs [7][]float64
	for y := miny; y < maxy && y < ny; y++ {
		bar.Advance(1)
		for band := 0; band < 7; band++ {
			rrs[band] = make([]float64, dx)
			if err = ds.RasterBand(band+1).IO(gdal.Read, minx, y, dx, 1, rrs[band], dx, 1, 0, 0); err != nil {
				return err
			}
		}
		var obs []Observation
		var skips []bool
		for x := minx; x < maxx; x++ {
			xx := make([]float64, 7)
			skip := false
			for band := 0; band < 7; band++ {
				v := rrs[band][x-minx]
				if math.IsNaN(v) {
					skip = true
				}
				xx[band] = v
			}
			xxt := data.Transform(xx, isL8)
			if len(xxt) != NVariables {
				return fmt.Errorf("expected %d input variables, found %d", NVariables, len(xxt))
			}
			var xxo Observation
			for i, v := range xxt {
				xxo[i] = v
			}
			obs = append(obs, xxo)
			skips = append(skips, skip)
		}
		res, err := model.Predict(obs)
		if err != nil {
			return err
		}
		row := make([]int8, len(res))
		for i := range res {
			if skips[i] {
				row[i] = 0
			} else {
				row[i] = int8(res[i] + 1)
			}
		}

		err = wds.IO(gdal.Write, minx, y, dx, 1, row, dx, 1, 0, 0)
		if err != nil {
			return err
		}
	}
	bar.Stop()
	return nil
}
