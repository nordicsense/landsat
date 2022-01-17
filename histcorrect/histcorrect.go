package main

import (
	"fmt"
	"github.com/nordicsense/landsat/dataset"
	"log"
	"math"
	"path"
	"strconv"
)

func HistCorrect(root, prefix string, options ...string) error {
	var (
		err error
		r   dataset.MultiBandReader
		w   dataset.MultiBandWriter
		box dataset.Box
		buf []float64
	)

	minScale := 0.04
	if len(options) > 0 {
		if minScale, err = strconv.ParseFloat(options[0], 64); err != nil {
			minScale = 0.04
		}
		options = options[1:]
	}

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
				if err = correctBand(buf, bands[i], minScale); err == nil {
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
		{Index: 1, Min: 0.065, Max: 0.12, Mods: []float64{0.085}, Vols: []float64{0.99}, MinSpread: 10},
		{Index: 2, Min: 0.03, Max: 0.12, Mods: []float64{0.045, 0.2}, Vols: []float64{0.065, 0.8}, MinSpread: 10},
		{Index: 3, Min: 0.01, Max: 0.11, Mods: []float64{0.025, 0.2}, Vols: []float64{0.05, 0.8}, MinSpread: 10},
		{Index: 4, Min: -0.1, Max: 0.33, Mods: []float64{0.02, 0.12}, Vols: []float64{0.06, 0.087}, MinSpread: 10},
		{Index: 5, Min: -0.1, Max: 0.21, Mods: []float64{0.003, 0.2}, Vols: []float64{0.11, 0.8}, MinSpread: 30},
		{Index: 6, Min: 90.0, Max: 165., Mods: []float64{130}, Vols: []float64{0.99}, MinSpread: 30},
		{Index: 7, Min: -0.1, Max: 0.14, Mods: []float64{0.002, 0.2}, Vols: []float64{0.05, 0.8}, MinSpread: 20},
	}
)

func correctBand(buf []float64, band Band, minScale float64) error {
	if len(band.Mods) > 2 {
		return fmt.Errorf("supported at most 2 mods, provided %d", len(band.Mods))
	}

	min, max, freq := histogram(buf)
	// mods, scales := mods2(min, max, rawhist, minScale, band)
	dens := freq2dens(freq)
	mods, vols := computeMods(min, max, dens, band.MinSpread)

	// if no mods found, then perform no correction
	if len(mods) == 0 {
		log.Printf("Found no mods for correction in band %d\n", band.Index)
		return nil
	}

	var factor, spread, center float64
	if len(band.Mods) == 2 && len(mods) == 2 {
		// scale to the center between two mods, spread centering there
		center = 0.5 * (band.Mods[0] + band.Mods[1])
		factor = center / (0.5 * (mods[0] + mods[1]))
		spread = (mods[1] - mods[0]) / (band.Mods[1] - band.Mods[0])
	} else if len(band.Mods) == 2 {
		// linear scaling to the second mode only
		center = band.Mods[1]
		factor = center / mods[0]
		spread = vols[0] / band.Vols[1]
	} else {
		// linear scaling to the first mode only, or just 1 mod
		center = band.Mods[0]
		factor = center / mods[0]
		spread = vols[0] / band.Vols[0]
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
		if !math.IsNaN(v) {
			buf[i] = correct(v)
		}
	}
	return nil
}
