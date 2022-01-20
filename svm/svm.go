package svm

import (
	"log"
	"math"
	"math/rand"

	"github.com/nordicsense/landsat/field"
	libSvm "github.com/nordicsense/libsvm-go"
)

const numBands = 5

func Process(data map[string]field.Records) {
	svs, _ := toSVs(data, 2000)
	p, err := libSvm.NewProblem(svs)
	if err != nil {
		log.Fatal(err)
	}

	// https://scikit-learn.org/stable/auto_examples/svm/plot_rbf_parameters.html
	par := &libSvm.Parameter{
		SvmType:     libSvm.C_SVC,
		KernelType:  libSvm.RBF,
		Gamma:       0.0002, // 1. / float64(numBands),
		Eps:         1e-7,
		C:           1000,
		NrWeight:    11,
		WeightLabel: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		Weight:      []float64{2.0, 2.0, 1.0, 1.0, 1.0, 1.0, 1.0, 3.0, 5.0, 3.0, 3.0},
		CacheSize:   2000,
		NumCPU:      4,
		QuietMode:   true,
	}
	m := libSvm.NewModel(par)

	log.Printf("training model with %d vectors\n", len(svs))
	err = m.Train(p)
	if err != nil {
		log.Fatal(err)
	}
	var matches, mismatches int
	log.Printf("validating %d cases\n", len(svs))
	for _, sv := range svs {
		if m.PredictVector(sv.Nodes) == sv.Label {
			matches++
		} else {
			mismatches++
		}
	}
	log.Printf("accuracy %f\n", float64(matches)/float64(matches+mismatches))
}

func toSVs(data map[string]field.Records, nSVs int) ([]libSvm.SV, map[string]float64) {
	clazzes := map[string]float64{
		"water_with_no_sediments":        1.,
		"water_with_sediments":           2.,
		"industrial_water":               3.,
		"wet_tailing_pond":               4.,
		"residential_area":               5.,
		"industrial_area":                6.,
		"old_burnt_area":                 7.,
		"tundra_stone_tundra":            8.,
		"tundra_undam_stone_with_lichen": 9.,
		"natural_undam_birch_forest_with_lichen_dwarf_shrub":    10.,
		"natural_undam_pine_forest_with_dwarf_shrub_and_lichen": 11.,
	}
	norm := normalizer(data)

	var res []libSvm.SV
	for clazz, label := range clazzes {
		for _, rr := range subsample(data[clazz], nSVs) {
			res = append(res, libSvm.NewDenseSV(label, norm(rr)...))
		}
	}
	return res, clazzes
}

func subsample(rrs field.Records, nSVs int) field.Records {
	if nSVs >= len(rrs) {
		res := make(field.Records, len(rrs))
		for i, rr := range rrs {
			res[i] = rr[:numBands] // FIXME dropping some bands
		}
		return res
	}
	idx := rand.Perm(len(rrs))
	res := make(field.Records, nSVs)
	for i := 0; i < nSVs; i++ {
		res[i] = rrs[idx[i]][:numBands] // FIXME dropping some bands
	}
	return res
}

func normalizer(data map[string]field.Records) func([]float64) []float64 {
	mins := make([]float64, numBands)
	maxs := make([]float64, numBands)
	for i := range mins {
		mins[i] = 1e16
	}
	for _, rrs := range data {
		for _, rr := range rrs {
			for i, v := range rr[:numBands] {
				if math.IsNaN(v) {
					continue
				}
				if v > maxs[i] {
					maxs[i] = v
				} else if v < mins[i] {
					mins[i] = v
				}
			}
		}
	}
	log.Println(mins)
	log.Println(maxs)
	return func(xx []float64) []float64 {
		res := make([]float64, numBands)
		for i, v := range xx {
			res[i] = (v-mins[i])/(maxs[i]-mins[i]) - 0.5
		}
		return res
	}
}
