package histcorrect

func freq2dens(freq []int) (dens []float64) {
	total := 0.
	for _, f := range freq {
		total += float64(f)
	}
	dens = make([]float64, len(freq))
	for i, f := range freq {
		dens[i] = float64(f) / float64(total)
	}
	return
}

func find1mod(min, max float64, dens []float64) float64 {
	mods, vols := getAllMods(min, max, dens)
	maxi := 0
	maxv := 0.
	for i, v := range vols {
		if v > maxv {
			maxv = v
			maxi = i
		}
	}
	return mods[maxi]
}

func find2mods(min, max float64, dens []float64, tv0, tv1 float64) ([]float64, []float64) {
	mx, vx := getAllMods(min, max, dens)

	// collect mods on the left to 60% of tv0
	i0 := 0
	v0 := 0.
	maxv0 := 0.
	for i := i0; i < len(vx) && v0 < 0.6*tv0; i++ {
		if vx[i] > maxv0 {
			maxv0 = vx[i]
			i0 = i
		}
		v0 += vx[i]
	}
	// collect mods on the right to 60% of tv1
	i1 := len(vx) - 1
	maxv1 := 0.
	v1 := 0.
	for i := i1; i >= 0 && v1 < 0.6*tv1; i-- {
		if vx[i] > maxv1 {
			maxv1 = vx[i]
			i1 = i
		}
		v1 += vx[i]
	}
	if i0 < i1 {
		// two distinct points
		mid := (mx[i0] + mx[i1]) * 0.5
		vols := []float64{0., 0.}
		mods := []float64{0., 0.}
		maxv0 = 0.
		maxv1 = 0.
		for i := 0; i < len(mx); i++ {
			if mx[i] <= mid {
				vols[0] += vx[i]
				if vx[i] > maxv0 {
					maxv0 = vx[i]
					mods[0] = mx[i]
				}
			} else {
				vols[1] += vx[i]
				if vx[i] > maxv1 {
					maxv1 = vx[i]
					mods[1] = mx[i]
				}
			}
		}
		return mods, vols
	} else {
		// assume single mod
		maxi := 0
		maxv := 0.
		total := 0.
		for i, v := range vx {
			if v > maxv {
				maxv = v
				maxi = i
			}
			total += v
		}
		return []float64{mx[maxi]}, []float64{total}
	}
}

func getAllMods(min, max float64, dens []float64) (mods, vols []float64) {
	type section struct{ start, max, end int }
	var (
		s       *section
		ss      []section
		growing bool
	)

	dlast := 0.0
	for i, d := range dens {
		if i == 0 {
			continue
		}
		if s == nil {
			if growing = d > 0 && d >= dens[i-1]; growing {
				dlast = dens[i-1]
				s = &section{start: i - 1, max: i, end: i}
			}
			continue
		}
		if growing {
			if d > 0 && d >= dlast {
				// update s if another value towards growth
				s.max = i
				s.end = i
				dlast = d
			} else if d > 0 {
				// turning point, update end, keep max
				s.end = i
				dlast = d
				growing = false
			}
		} else {
			if d > 0 && d >= dlast {
				// new minimum: add s as is, start new s
				ss = append(ss, *s)
				s = &section{start: i, max: i, end: i}
				dlast = d
				growing = true
			} else if d > 0 {
				// decreasing, update end for s
				s.end = i
				dlast = d
			}
		}
		if i == len(dens)-1 && !growing && s != nil {
			ss = append(ss, *s)
		}
	}

	delta := (max - min) / float64(len(dens))
	for _, v := range ss {
		mod := min + delta*float64(v.max+1)
		var vol float64
		for _, d := range dens[v.start:(v.end + 1)] {
			vol += d
		}
		mods = append(mods, mod)
		vols = append(vols, vol)
	}
	return
}
