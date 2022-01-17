package main

const minVol = 0.05

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

// min, max - data range, dens -- density histogram - density threshold
func computeMods(min, max float64, dens []float64, minSpread int) (mods, vols []float64) {
	type section struct{ start, max, end int }
	var (
		s       *section
		ss      []section
		growing bool
	)

	appendSection := func(s section) {
		if prev := len(ss) - 1; prev >= 0 && s.max-ss[prev].max < minSpread {
			ss[prev].end = s.end
			if dens[s.max] > dens[ss[prev].max] {
				ss[prev].max = s.max
			}
		} else {
			ss = append(ss, s)
		}
	}

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
				appendSection(*s)
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
			appendSection(*s)
		}
	}

	delta := (max - min) / float64(len(dens))
	for _, v := range ss {
		mod := min + delta*float64(v.max+1)
		var vol float64
		for _, d := range dens[v.start:(v.end + 1)] {
			vol += d
		}
		if vol >= minVol {
			mods = append(mods, mod)
			vols = append(vols, vol)
		}
	}
	return
}
