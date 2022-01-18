package hist

import (
	"math"
	"sort"
)

func Compute(buf []float64) (min, max float64, hist []int) {
	var data []float64
	for _, v := range buf {
		if !math.IsNaN(v) {
			data = append(data, v)
		}
	}
	hist = make([]int, 100)
	if len(data) < 1 {
		return
	}
	sort.Float64s(data)

	min = data[0]
	max = data[len(data)-1]
	delta := (max - min) / 100.
	total := 0
	for i, j := 0, 0; i < 100; i++ {
		for ; j < len(data) && data[j] < min+delta*float64(i+1); j++ {
			hist[i] += 1
			total++
		}
	}

	// cutoff the tails of 1% area left and right

	ileft := 0
	for volleft := 0; volleft < total/100 && ileft < 100; ileft++ {
		volleft += hist[ileft]
	}
	iright := 99
	for volright := 0; volright < total/100 && iright >= 0; iright-- {
		volright += hist[iright]
	}
	max = min + delta*float64(iright+1) // uses previous min
	min = min + delta*float64(ileft+1)

	// recompute hist:
	data = nil
	for _, v := range buf {
		if !math.IsNaN(v) && v >= min && v <= max {
			data = append(data, v)
		}
	}
	hist = make([]int, 100)
	if len(data) < 1 {
		return
	}
	sort.Float64s(data)

	min = data[0]
	max = data[len(data)-1]
	delta = (max - min) / 100.
	total = 0
	for i, j := 0, 0; i < 100; i++ {
		for ; j < len(data) && data[j] < min+delta*float64(i+1); j++ {
			hist[i] += 1
			total++
		}
	}
	return
}
