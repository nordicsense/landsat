package data

var (
	mins = []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
	maxs = []float64{0.2, 0.2, 0.2, 0.5, 0.4, 0.3}
)

const NVars = 10

var (
	Clazzes = []string{"band1", "band2", "band3", "band4", "band5", "band7", "ndvi", "nbr", "nbr2", "ndwi"}
)

func Transform(data []float64, landsatId int) []float64 {
	res := make([]float64, NVars)
	switch landsatId {
	case 5:
		copy(res, data[0:5])
		res[5] = data[6]
		res[4] = (res[4]+0.005)*1.2 - 0.005
	case 8:
		// correct for L8, where the 1st classic band starts at 2nd pos and band 6 is missing
		copy(res, data[1:6])
		res[4] /= 65535.
		res[4] = (res[4]-0.14)*2. + 0.14
	default:
		copy(res, data[0:5])
		res[5] = data[6]
	}

	res[6] = (res[3]-res[2])/(res[3]+res[2])/2. + 0.5 // NDVI (scaled to [0,1])
	res[7] = (res[3]-res[5])/(res[3]+res[5])/2. + 0.5 // NBR (scaled to [0,1])
	res[8] = (res[4]-res[5])/(res[4]+res[5])/2. + 0.5 // NBR2 (scaled to [0,1])
	res[9] = (res[2]-res[4])/(res[2]+res[4])/2. + 0.5 // NDWI (scaled to [0,1])

	for i := 0; i < 6; i++ {
		res[i] = (res[i] - mins[i]) / (maxs[i] - mins[i])
	}
	return res
}
