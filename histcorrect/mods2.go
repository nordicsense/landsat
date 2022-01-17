package main

import "math"

func mods2(min, max float64, rawhist []int, cutoff float64, band Band) (mods []float64, scales []float64) {
	delta := (max - min) / float64(len(rawhist))
	total := 0.
	for _, v := range rawhist {
		total += float64(v)
	}

	growing := false
	lastscale := math.NaN()
	for i, v := range rawhist {
		if i == 0 {
			continue
		}
		scale := float64(v) / total
		if i == 1 {
			growing = scale >= float64(rawhist[0])/total
		} else if growing {
			if scale < lastscale {
				growing = false
				if scale > cutoff {
					mods = append(mods, min+delta*float64(i+1))
					if i > 1 {
						scale = (float64(rawhist[i-2])/total + lastscale + scale) / 3.0
					} else {
						scale = (lastscale + scale) / 2.0
					}
					scales = append(scales, scale)
				}
			}
		} else {
			if scale > lastscale {
				growing = true
			}
		}
		lastscale = scale
	}

	if len(mods) > 0 {
		mods, scales = filterMods(mods, scales, len(band.Mods))
	}
	return
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
