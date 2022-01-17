package main

import (
	"errors"
	"fmt"
	"github.com/nordicsense/landsat/dataset"
	"math"
	"path"
)

type Band struct {
	Min, Max float64
	Mods     []float64
	Scales   []float64
}

const minScale = 0.01

var (
	// DO NOT use more than 2 mods, at the moment unsupported
	bands = []Band{
		{Min: 0.065, Max: 0.12, Mods: []float64{0.085}, Scales: []float64{0.12}},
		{Min: 0.03, Max: 0.12, Mods: []float64{0.045, 0.07}, Scales: []float64{0.06, 0.12}},
		{Min: 0.01, Max: 0.11, Mods: []float64{0.025, 0.04}, Scales: []float64{0.05, 0.11}},
		{Min: -0.1, Max: 0.33, Mods: []float64{0.02, 0.2}, Scales: []float64{0.06, 0.025}},
		{Min: -0.1, Max: 0.21, Mods: []float64{0.003, 0.13}, Scales: []float64{0.06, 0.03}},
		{Min: 90.0, Max: 165., Mods: []float64{129}, Scales: []float64{0.15}},
		{Min: -0.1, Max: 0.14, Mods: []float64{0.02, 0.48}, Scales: []float64{0.06, 0.06}},
	}
)

func HistCorrect(root, prefix string, options ...string) error {
	var (
		err error
		r   dataset.MultiBandReader
		w   dataset.MultiBandWriter
		box dataset.Box
		buf []float64
	)

	if r, err = dataset.OpenMultiBand(path.Join(root, prefix+".tiff")); err == nil {
		defer r.Close()
		ip := r.ImageParams()
		box = dataset.Box{0, 0, ip.XSize(), ip.YSize()}
		if w, err = dataset.NewMultiBand(path.Join(root, prefix+"_histcorr.tiff"), dataset.GTiff, r.Bands(), ip, options...); err == nil {
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

func correctBand(buf []float64, band Band) error {
	if len(band.Mods) > 2 {
		return fmt.Errorf("supported at max 2 mods, provided %d", len(band.Mods))
	}

	_, _, rawhist := histogram(buf)
	total := 0.
	for _, v := range rawhist {
		total += float64(v)
	}
	var (
		hist   []float64
		mods   []float64
		scales []float64
	)
	last := math.NaN()
	for _, v := range rawhist {
		hv := float64(v) / total
		if math.IsNaN(last) || hv >= last {
			last = hv
		} else if scale := hist[len(hist)-1]; scale > minScale {
			mods = append(mods, last)
			scale := hist[len(hist)-1] + hv
			if len(hist)-2 > 0 {
				scale = (scale + hist[len(hist)-2]) / 3.
			} else {
				scale /= 2.
			}
			scales = append(scales, scale)
			last = math.NaN()
		}
		hist = append(hist, hv)
	}

	// if no mods found, then perform no correction
	if len(mods) == 0 {
		return errors.New("found no mods for correction")
	}

	mods, scales = filterMods(mods, scales, len(band.Mods))

	var factor, spread, center float64
	if len(band.Mods) == 2 && len(mods) == 2 {
		// scale to the center between two mods, spread centering there
		center = 0.5 * (band.Mods[0] + band.Mods[1])
		factor = center / (0.5 * (mods[0] + mods[1]))
		spread = (band.Mods[1] - band.Mods[0]) / (mods[1] - mods[0])
	} else {
		// linear scaling to the first mode only
		center = band.Mods[0]
		factor = center / mods[0]
		spread = band.Scales[0] / scales[0]
	}
	correct := func(v float64) float64 {
		// scale to mod, spread centering on the mod
		res := (v*factor-center)*spread + center
		if res < band.Min {
			res = band.Min
		} else if res > band.Max {
			res = math.NaN()
		}
		return res
	}

	for i, v := range buf {
		buf[i] = correct(v)
	}
	return nil
}

func filterMods(mods, scales []float64, n int) ([]float64, []float64) {
	if len(mods) <= n {
		return mods, scales
	}
	maxi := 0
	maxmod := mods[0]
	maxscale := scales[0]
	for i := 1; i < len(mods); i++ {
		if scales[i] > maxscale {
			maxmod = mods[i]
			maxscale = scales[i]
			maxi = i
		}
	}
	if n == 1 {
		return []float64{maxmod}, []float64{maxscale}
	}
	// n == 2
	secondmod := math.NaN()
	secondscale := math.NaN()
	secondi := -1
	for i := 0; i < len(mods); i++ {
		if i == maxi {
			continue
		}
		if math.IsNaN(secondscale) || scales[i] > secondscale {
			secondmod = mods[i]
			secondscale = scales[i]
			secondi = i
		}
	}
	if maxi > secondi {
		return []float64{secondmod, maxmod}, []float64{secondscale, maxscale}
	}
	return []float64{maxmod, secondmod}, []float64{maxscale, secondscale}
}
