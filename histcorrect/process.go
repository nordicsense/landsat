package histcorrect

import (
	"fmt"
	"github.com/nordicsense/landsat/dataset"
	"github.com/nordicsense/landsat/hist"
	"log"
	"math"
	"path"
	"strings"
)

func Process(fName, pathOut string, options ...string) error {
	var (
		err error
		r   dataset.MultiBandReader
		w   dataset.MultiBandWriter
		box dataset.Box
		buf []float64
	)

	if r, err = dataset.OpenMultiBand(fName); err == nil {
		defer r.Close()
		ip := r.ImageParams()
		box = dataset.Box{0, 0, ip.XSize(), ip.YSize()}
		fNameOut := path.Join(pathOut, strings.Replace(path.Base(fName), path.Ext(fName), "", 1)+"_histcorr.tiff")
		if w, err = dataset.NewMultiBand(fNameOut, dataset.GTiff, r.Bands(), ip, options...); err == nil {
			defer w.Close()
		}
	}
	for i := 0; err == nil && i < r.Bands(); i++ {
		br := r.Reader(i + 1)
		bw := w.Writer(i + 1)

		if err = bw.SetRasterParams(br.RasterParams()); err == nil {
			if buf, err = br.ReadBlock(0, 0, box); err == nil {
				if err = correctBand(buf, bands[i]); err == nil {
					err = bw.WriteBlock(0, 0, box, buf)
				}
			}
		}
	}
	return err
}

type Band struct {
	Index     int
	Min, Max  float64
	Mods      []float64
	Vols      []float64
	MinSpread int
}

var (
	// DO NOT use more than 2 mods, at the moment unsupported
	bands = []Band{
		{Index: 1, Min: 0.06, Max: 0.12, Mods: []float64{0.085}, Vols: []float64{100.}},
		{Index: 2, Min: 0.03, Max: 0.12, Mods: []float64{0.045, 0.065}, Vols: []float64{0.2, 0.8}},
		{Index: 3, Min: 0.01, Max: 0.11, Mods: []float64{0.025, 0.05}, Vols: []float64{0.2, 0.8}},
		{Index: 4, Min: 0.0, Max: 0.33, Mods: []float64{0.02, 0.2}, Vols: []float64{0.2, 0.8}},
		{Index: 5, Min: -0.1, Max: 0.21, Mods: []float64{0.003, 0.12}, Vols: []float64{0.2, 0.8}},
		{Index: 6, Min: 90.0, Max: 165., Mods: []float64{130}, Vols: []float64{100.}},
		{Index: 7, Min: -0.01, Max: 0.14, Mods: []float64{0.002, 0.05}, Vols: []float64{0.2, 0.8}},
	}
)

func correctBand(buf []float64, band Band) error {
	min, max, freq := hist.Compute(buf)
	dens := freq2dens(freq)

	var mods []float64
	if len(band.Vols) == 1 {
		mods = []float64{find1mod(min, max, dens)}
	} else if len(band.Vols) == 2 {
		mods, _ = find2mods(min, max, dens, band.Vols[0], band.Vols[1])
	} else {
		return fmt.Errorf("supported at most 2 mods, provided %d", len(band.Mods))
	}

	// if no mods found, then perform no correction
	if len(mods) == 0 {
		log.Printf("Found no mods for correction in band %d\n", band.Index)
		return nil
	}

	var correct func(v float64) float64

	if len(band.Mods) == 2 && len(mods) == 2 {
		// scale to the center between two mods, spread centering there
		center := 0.5 * (band.Mods[0] + band.Mods[1])
		factor := center / (0.5 * (mods[0] + mods[1]))
		spread := (band.Mods[1] - band.Mods[0]) / (mods[1] - mods[0]) / factor
		correct = func(v float64) float64 {
			if math.IsNaN(v) || v < min || v > max {
				return math.NaN()
			}
			// scale to mod, spread centering on the mod
			res := (v*factor-center)*spread + center
			if res < band.Min {
				res = band.Min
			} else if res > band.Max {
				res = math.NaN()
			}
			return res
		}
	} else if len(band.Mods) == 2 {
		// linear scaling to the second mode only
		factor := band.Mods[1] / mods[0]
		correct = func(v float64) float64 {
			if math.IsNaN(v) || v < min || v > max {
				return math.NaN()
			}
			res := v * factor
			if res < band.Min {
				res = band.Min
			} else if res > band.Max {
				res = math.NaN()
			}
			return res
		}
	} else {
		// linear scaling to the first mode only, or just 1 mod
		factor := band.Mods[0] / mods[0]
		correct = func(v float64) float64 {
			res := v * factor
			if res < band.Min {
				res = band.Min
			} else if res > band.Max {
				res = math.NaN()
			}
			return res
		}
	}

	for i, v := range buf {
		if !math.IsNaN(v) {
			buf[i] = correct(v)
		}
	}
	return nil
}
