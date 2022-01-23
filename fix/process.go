package fix

import (
	"errors"
	"fmt"
	"github.com/nordicsense/landsat/dataset"
	"math"
	"path"
	"strings"
)

type Band struct {
	Index    int
	Min, Max float64
	Target   interface{}
	Images   map[string]interface{}
}

type SingleMod struct {
	X, Scale float64
}

type DualMod struct {
	X0, X1, Scale float64 // Further Scale centering on 2nd mod, applied first (mods adjusted)
}

var bands = []Band{
	{
		Index: 1, Min: 0.07, Max: 0.14, Target: SingleMod{X: 0.092, Scale: 1.0},
		Images: map[string]interface{}{
			"LT05_L1TP_187012_20050709": SingleMod{X: 0.095, Scale: 1.},
			"LT05_L1TP_187013_20050709": SingleMod{X: 0.088, Scale: 1.},
			"LT05_L1TP_188012_19850709": SingleMod{X: 0.092, Scale: 1.}, // ref
			"LT05_L1TP_188012_19900723": SingleMod{X: 0.096, Scale: 0.8},
			"LT05_L1TP_188013_19850709": SingleMod{X: 0.090, Scale: 1.},
			"LT05_L1TP_188013_19900723": SingleMod{X: 0.084, Scale: 1.},
			"LE07_L1TP_186012_20000728": SingleMod{X: 0.092, Scale: 0.6},
			"LE07_L1TP_188012_20000726": SingleMod{X: 0.086, Scale: 0.95},
			"LE07_L1TP_195011_20000727": SingleMod{X: 0.089, Scale: 0.95},
			"LE07_L1TP_195012_20000727": SingleMod{X: 0.085, Scale: 1.1},
			// "LT05_L1TP_187012_20050709": SingleMod{X: , Scale: 1.},
			"LT05_L1TP_188012_19860728": SingleMod{X: 0.099, Scale: 1.},
			"LT05_L1TP_190011_20090725": SingleMod{X: 0.097, Scale: 0.7},
			"LT05_L1TP_190012_19930713": SingleMod{X: 0.097, Scale: 1.05},
		},
	},
	{
		Index: 2, Min: 0.04, Max: 0.14, Target: DualMod{X0: 0.05, X1: 0.08, Scale: 1.0},
		Images: map[string]interface{}{
			"LT05_L1TP_187012_20050709": DualMod{X0: 0.05, X1: 0.08, Scale: 1.0}, // ref
			"LT05_L1TP_187013_20050709": DualMod{X0: 0.05, X1: 0.07, Scale: 0.9},
			"LT05_L1TP_188012_19850709": DualMod{X0: 0.055, X1: 0.088, Scale: 0.75},
			"LT05_L1TP_188012_19900723": DualMod{X0: 0.05, X1: 0.084, Scale: 1.0},
			"LT05_L1TP_188013_19850709": DualMod{X0: 0.05, X1: 0.075, Scale: 1.0},
			"LT05_L1TP_188013_19900723": DualMod{X0: 0.047, X1: 0.08, Scale: 0.75},
			"LE07_L1TP_186012_20000728": DualMod{X0: 0.048, X1: 0.075, Scale: 0.65},
			"LE07_L1TP_188012_20000726": DualMod{X0: 0.048, X1: 0.07, Scale: 0.75},
			"LE07_L1TP_195011_20000727": DualMod{X0: 0.048, X1: 0.075, Scale: 0.75},
			"LE07_L1TP_195012_20000727": DualMod{X0: 0.048, X1: 0.07, Scale: 0.85},
			// "LT05_L1TP_187012_20050709": SingleMod{X: , Scale: 1.},
			"LT05_L1TP_188012_19860728": DualMod{X0: 0.05, X1: 0.085, Scale: 0.95},
			"LT05_L1TP_190011_20090725": DualMod{X0: 0.055, X1: 0.082, Scale: 0.8},
			"LT05_L1TP_190012_19930713": DualMod{X0: 0.058, X1: 0.083, Scale: 1.0},
		},
	},
	{
		Index: 3, Min: 0.02, Max: 0.14, Target: SingleMod{X: 0.065, Scale: 1.0},
		Images: map[string]interface{}{
			"LT05_L1TP_187012_20050709": SingleMod{X: 0.065, Scale: 1.},   //ref
			"LT05_L1TP_187013_20050709": SingleMod{X: 0.053, Scale: 0.85}, // bad
			"LT05_L1TP_188012_19850709": SingleMod{X: 0.065, Scale: 1.1},
			"LT05_L1TP_188012_19900723": SingleMod{X: 0.067, Scale: 1.},
			"LT05_L1TP_188013_19850709": SingleMod{X: 0.055, Scale: 1.1},
			"LT05_L1TP_188013_19900723": SingleMod{X: 0.057, Scale: 1.1},
			"LE07_L1TP_186012_20000728": SingleMod{X: 0.065, Scale: 0.6},
			"LE07_L1TP_188012_20000726": SingleMod{X: 0.047, Scale: 0.85},
			"LE07_L1TP_195011_20000727": SingleMod{X: 0.05, Scale: 0.8},
			"LE07_L1TP_195012_20000727": SingleMod{X: 0.047, Scale: 1.0},
			// "LT05_L1TP_187012_20050709": SingleMod{X: , Scale: 1.},
			"LT05_L1TP_188012_19860728": SingleMod{X: 0.065, Scale: 1.},
			"LT05_L1TP_190011_20090725": SingleMod{X: 0.069, Scale: 0.85},
			"LT05_L1TP_190012_19930713": SingleMod{X: 0.065, Scale: 1.1},
		},
	},
	{
		Index: 4, Min: 0.01, Max: 0.33, Target: DualMod{X0: 0.035, X1: 0.22, Scale: 1.},
		Images: map[string]interface{}{
			"LT05_L1TP_187012_20050709": DualMod{X0: 0.024, X1: 0.235, Scale: 1.},
			"LT05_L1TP_187013_20050709": DualMod{X0: 0.024, X1: 0.22, Scale: 0.8},
			"LT05_L1TP_188012_19850709": DualMod{X0: 0.035, X1: 0.22, Scale: 1.}, //ref
			"LT05_L1TP_188012_19900723": DualMod{X0: 0.035, X1: 0.235, Scale: 0.9},
			"LT05_L1TP_188013_19850709": DualMod{X0: 0.024, X1: 0.2, Scale: 0.9},
			"LT05_L1TP_188013_19900723": DualMod{X0: 0.024, X1: 0.22, Scale: 0.9},
			"LE07_L1TP_186012_20000728": DualMod{X0: 0.024, X1: 0.22, Scale: 0.9},
			"LE07_L1TP_188012_20000726": DualMod{X0: 0.024, X1: 0.22, Scale: 0.8},
			"LE07_L1TP_195011_20000727": DualMod{X0: 0.024, X1: 0.22, Scale: 0.85},
			"LE07_L1TP_195012_20000727": DualMod{X0: 0.024, X1: 0.22, Scale: 0.9},
			// "LT05_L1TP_187012_20050709": SingleMod{X: , Scale: 1.},
			"LT05_L1TP_188012_19860728": DualMod{X0: 0.035, X1: 0.21, Scale: 1.},
			"LT05_L1TP_190011_20090725": DualMod{X0: 0.035, X1: 0.22, Scale: 1.1},
			"LT05_L1TP_190012_19930713": DualMod{X0: 0.035, X1: 0.18, Scale: 0.8},
		},
	},
	{
		Index: 5, Min: 0.0, Max: 0.26, Target: DualMod{X0: 0.0055, X1: 0.155, Scale: 1.},
		Images: map[string]interface{}{
			"LT05_L1TP_187012_20050709": DualMod{X0: 0.0055, X1: 0.155, Scale: 1.},
			"LT05_L1TP_187013_20050709": DualMod{X0: 0.0055, X1: 0.12, Scale: 1.},
			"LT05_L1TP_188012_19850709": DualMod{X0: 0.0055, X1: 0.155, Scale: 1.}, //ref
			"LT05_L1TP_188012_19900723": DualMod{X0: 0.013, X1: 0.147, Scale: 1.},
			"LT05_L1TP_188013_19850709": DualMod{X0: 0.0055, X1: 0.145, Scale: 0.95},
			"LT05_L1TP_188013_19900723": DualMod{X0: 0.005, X1: 0.135, Scale: 1.},
			"LE07_L1TP_186012_20000728": DualMod{X0: 0.0055, X1: 0.18, Scale: 1.},
			"LE07_L1TP_188012_20000726": DualMod{X0: 0.0055, X1: 0.145, Scale: 1.},
			"LE07_L1TP_195011_20000727": DualMod{X0: 0.0065, X1: 0.2, Scale: 0.8},
			"LE07_L1TP_195012_20000727": DualMod{X0: 0.009, X1: 0.155, Scale: 1.},
			// "LT05_L1TP_187012_20050709": SingleMod{X: , Scale: 1.},
			"LT05_L1TP_188012_19860728": DualMod{X0: 0.0055, X1: 0.14, Scale: 1.1},
			"LT05_L1TP_190011_20090725": DualMod{X0: 0.007, X1: 0.15, Scale: 0.85},
			"LT05_L1TP_190012_19930713": DualMod{X0: 0.0055, X1: 0.14, Scale: 1.2},
		},
	},
	{
		Index: 6, Min: 90.0, Max: 165., Target: SingleMod{X: 140., Scale: 1.0},
		Images: map[string]interface{}{
			"LT05_L1TP_187012_20050709": SingleMod{X: 142., Scale: 1.0},
			"LT05_L1TP_187013_20050709": SingleMod{X: 138., Scale: 1.0},
			"LT05_L1TP_188012_19850709": SingleMod{X: 140., Scale: 1.0}, // ref
			"LT05_L1TP_188012_19900723": SingleMod{X: 120., Scale: 0.9},
			"LT05_L1TP_188013_19850709": SingleMod{X: 133., Scale: 0.8},
			"LT05_L1TP_188013_19900723": SingleMod{X: 126., Scale: 0.8},
			"LE07_L1TP_186012_20000728": SingleMod{X: 133., Scale: 1.15},
			"LE07_L1TP_188012_20000726": SingleMod{X: 120., Scale: 1.},
			"LE07_L1TP_195011_20000727": SingleMod{X: 135., Scale: 1.},
			"LE07_L1TP_195012_20000727": SingleMod{X: 133., Scale: 1.15},
			// "LT05_L1TP_187012_20050709": SingleMod{X: , Scale: 1.},
			"LT05_L1TP_188012_19860728": SingleMod{X: 135., Scale: 1.},
			"LT05_L1TP_190011_20090725": SingleMod{X: 127, Scale: 0.65},
			"LT05_L1TP_190012_19930713": SingleMod{X: 135, Scale: 1.},
		},
	},
	{
		Index: 7, Min: 0.0, Max: 0.15, Target: SingleMod{X: 0.07, Scale: 1.0},
		Images: map[string]interface{}{
			"LT05_L1TP_187012_20050709": SingleMod{X: 0.07, Scale: 1.0},
			"LT05_L1TP_187013_20050709": SingleMod{X: 0.06, Scale: 1.0},
			"LT05_L1TP_188012_19850709": SingleMod{X: 0.07, Scale: 1.0},
			"LT05_L1TP_188012_19900723": SingleMod{X: 0.07, Scale: 1.0},
			"LT05_L1TP_188013_19850709": SingleMod{X: 0.07, Scale: 1.0},
			"LT05_L1TP_188013_19900723": SingleMod{X: 0.065, Scale: 1.0},
			"LE07_L1TP_186012_20000728": SingleMod{X: 0.07, Scale: 0.9},
			"LE07_L1TP_188012_20000726": SingleMod{X: 0.07, Scale: 0.9},
			"LE07_L1TP_195011_20000727": SingleMod{X: 0.07, Scale: 1.0},
			"LE07_L1TP_195012_20000727": SingleMod{X: 0.07, Scale: 1.0},
			// "LT05_L1TP_187012_20050709": SingleMod{X: , Scale: 1.},
			"LT05_L1TP_188012_19860728": SingleMod{X: 0.07, Scale: 1.},
			"LT05_L1TP_190011_20090725": SingleMod{X: 0.07, Scale: 0.9},
			"LT05_L1TP_190012_19930713": SingleMod{X: 0.07, Scale: 1.1},
		},
	},
}

func Process(fName, pathOut string, options ...string) error {
	var (
		err error
		r   dataset.MultiBandReader
		w   dataset.MultiBandWriter
		box dataset.Box
		buf []float64
	)
	img := path.Base(fName)[:25]
	if _, ok := bands[0].Images[img]; !ok {
		fmt.Printf("Ignoring %s as there are no parameters to apply fixes", fName)
		return nil
	}

	if r, err = dataset.OpenMultiBand(fName); err == nil {
		defer r.Close()
		ip := r.ImageParams()
		box = dataset.Box{0, 0, ip.XSize(), ip.YSize()}
		fNameOut := path.Join(pathOut, strings.Replace(path.Base(fName), path.Ext(fName), "", 1)+"_fix.tiff")
		if w, err = dataset.NewMultiBand(fNameOut, dataset.GTiff, r.Bands(), ip, options...); err == nil {
			defer w.Close()
		}
	}
	for i := 0; err == nil && i < r.Bands(); i++ {
		br := r.Reader(i + 1)
		bw := w.Writer(i + 1)
		if err = bw.SetRasterParams(br.RasterParams()); err == nil {
			if buf, err = br.ReadBlock(0, 0, box); err == nil {
				band := bands[i]
				par := band.Images[img]
				if t, ok := band.Target.(SingleMod); ok {
					correctSingleMod(buf, band.Min, band.Max, par.(SingleMod), t)
				} else if t, ok := band.Target.(DualMod); ok {
					correctDualMod(buf, band.Min, band.Max, par.(DualMod), t)
				} else {
					err = errors.New("unknown config")
				}
				if err == nil {
					err = bw.WriteBlock(0, 0, box, buf)
				}
			}
		}
	}
	return err
}

func correctSingleMod(buf []float64, min, max float64, mod, target SingleMod) {
	center := target.X
	scale2Center := center / mod.X
	scaleAtCenter := mod.Scale
	correct := func(v float64) float64 {
		if math.IsNaN(v) {
			return math.NaN()
		}
		res := (v*scale2Center-center)*scaleAtCenter + center
		if res < min {
			res = min
		} else if res > max {
			res = math.NaN()
		}
		return res
	}
	for i, v := range buf {
		if !math.IsNaN(v) {
			buf[i] = correct(v)
		}
	}
}

func correctDualMod(buf []float64, min, max float64, mods, target DualMod) {
	center := target.X1
	scale2Center := center / mods.X1
	scaleAtCenter := mods.Scale
	spreadFromCenter := (target.X1 - target.X0) / (mods.X1 - mods.X0*scale2Center)
	correct := func(v float64) float64 {
		if math.IsNaN(v) {
			return math.NaN()
		}
		res := (v*scale2Center-center)*spreadFromCenter*scaleAtCenter + center
		if res < min {
			res = min
		} else if res > max {
			res = math.NaN()
		}
		return res
	}
	for i, v := range buf {
		if !math.IsNaN(v) {
			buf[i] = correct(v)
		}
	}
}
