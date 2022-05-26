package data

var (
	mins = []float64{0.0, 0.0, 0.0, 0.0, 0.0}
	maxs = []float64{0.2, 0.2, 0.2, 0.5, 0.3}
)

const NVars = 7

var (
	Clazzes = []string{"band1", "band2", "band3", "band4", "band7", "ndvi", "nbr"}
)

func Transform(data []float64, landsatId int) []float64 {
	res := make([]float64, NVars)
	switch landsatId {
	case 8:
		copy(res, data[1:5])
		res[4] = data[6]
	default:
		copy(res, data[0:4])
		res[4] = data[6]
	}

	res[5] = (res[3]-res[2])/(res[3]+res[2])/2. + 0.5 // NDVI (scaled to [0,1])
	res[6] = (res[3]-res[4])/(res[3]+res[4])/2. + 0.5 // NBR (scaled to [0,1])

	for i := 0; i < 5; i++ {
		res[i] = (res[i] - mins[i]) / (maxs[i] - mins[i])
	}
	return res
}
