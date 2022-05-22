package data

/*
2022/05/22 16:50:51 col 0: 0.01=0.076054, 0.99=0.503638
2022/05/22 16:50:51 col 1: 0.01=0.042776, 0.99=0.547294
2022/05/22 16:50:51 col 2: 0.01=0.022875, 0.99=0.557613
2022/05/22 16:50:51 col 3: 0.01=0.014935, 0.99=0.556160
2022/05/22 16:50:51 col 4: 0.01=0.000222, 0.99=0.445470
2022/05/22 16:50:51 col 5: 0.01=-0.001106, 0.99=0.298056
*/

var (
	mins = []float64{0.076054, 0.042776, 0.022875, 0.014935, 0.000222, -0.001106}
	maxs = []float64{0.503638, 0.547294, 0.557613, 0.556160, 0.445470, 0.298056}
)

const n = 10

func Transform(data []float64, isL8 bool) []float64 {
	res := make([]float64, n)
	if isL8 {
		// correct for L8, where the 1st classic band starts at 2nd pos and band 6 is missing
		copy(res, data[1:6])
	} else {
		copy(res, data[0:5])
	}
	res[5] = data[6]

	res[6] = (res[3]-res[2])/(res[3]+res[2])/2. + 0.5 // NDVI (scaled to [0,1])
	res[7] = (res[3]-res[6])/(res[3]+res[6])/2. + 0.5 // NBR (scaled to [0,1])
	res[8] = (res[4]-res[6])/(res[4]+res[6])/2. + 0.5 // NBR2 (scaled to [0,1])
	res[9] = (res[3]-res[4])/(res[3]+res[4])/2. + 0.5 // NDWI (scaled to [0,1])

	for i := 0; i < 6; i++ {
		res[i] = (res[i] - mins[i]) / (maxs[i] - mins[i])
	}
	for i := 0; i < n; i++ {
		if res[i] < 0. {
			res[i] = 0.0
		} else if res[i] > 1.0 {
			res[i] = 1.0
		}
	}
	return res
}
