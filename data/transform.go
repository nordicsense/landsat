package data

const NVars = 10

var (
	Clazzes = []string{"band1", "band2", "band3", "band4", "band5", "band7", "ndvi", "nbr", "nbr2", "ndwi"}
)

func Transform(data []float64, landsatId int) []float64 {
	res := make([]float64, NVars)
	switch landsatId {
	case 8:
		copy(res, data[1:7])
	default:
		copy(res, data[0:5])
		res[5] = data[6]
	}

	res[6] = (res[3]-res[2])/(res[3]+res[2])/2. + 0.5 // NDVI (scaled to [0,1])
	res[7] = (res[3]-res[5])/(res[3]+res[5])/2. + 0.5 // NBR (scaled to [0,1])
	res[8] = (res[4]-res[5])/(res[4]+res[5])/2. + 0.5 // NBR2 (scaled to [0,1])
	res[9] = (res[2]-res[4])/(res[2]+res[4])/2. + 0.5 // NDWI (scaled to [0,1])

	return res
}
