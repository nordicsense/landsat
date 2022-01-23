package svm

import (
	"log"
	"math"
	"math/rand"

	"github.com/nordicsense/landsat/field"
	libSvm "github.com/nordicsense/libsvm-go"
)

const numBands = 7

var clazzes = map[string]float64{
	"white":    1.,
	"infra":    2.,
	"rock":     3.,
	"damaged":  4.,
	"fire":     5.,
	"birch":    6.,
	"pine":     7.,
	"spruce":   8.,
	"tundra":   9.,
	"water":    10.,
	"polluted": 11.,
	"wetland":  12.,
	"grass":    13.,
}

func Process(data []field.Record) {
	svs, _ := toSVs(data, 2000)
	p, err := libSvm.NewProblem(svs)
	if err != nil {
		log.Fatal(err)
	}

	// https://scikit-learn.org/stable/auto_examples/svm/plot_rbf_parameters.html
	par := &libSvm.Parameter{
		SvmType:     libSvm.C_SVC,
		KernelType:  libSvm.RBF,
		Gamma:       0.00001, //1. / float64(numBands),
		Eps:         1e-6,
		C:           1000,
		NrWeight:    13,
		WeightLabel: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
		Weight:      []float64{2.0, 0.2, 1.0, 1.0, 0.1, 2.0, 3.0, 3.0, 4.0, 3.0, 1.0, 2.0, 1.0},
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

func toSVs(rrs []field.Record, nSVs int) ([]libSvm.SV, map[string]float64) {
	norm := normalizer(rrs)

	xx := make(map[string][][]float64)
	for _, rr := range rrs {
		xx[rr.Clazz] = append(xx[rr.Clazz], rr.Bands)
	}

	var res []libSvm.SV
	for clazz, x := range xx {
		for _, rr := range subsample(x, nSVs) {
			res = append(res, libSvm.NewDenseSV(clazzes[clazz], norm(rr)...))
		}
	}
	return res, clazzes
}

func subsample(rrs [][]float64, nSVs int) [][]float64 {
	if nSVs >= len(rrs) {
		res := make([][]float64, len(rrs))
		for i, rr := range rrs {
			res[i] = rr[:numBands] // FIXME dropping some bands
		}
		return res
	}
	idx := rand.Perm(len(rrs))
	res := make([][]float64, nSVs)
	for i := 0; i < nSVs; i++ {
		res[i] = rrs[idx[i]][:numBands] // FIXME dropping some bands
	}
	return res
}

func normalizer(data []field.Record) func([]float64) []float64 {
	mins := make([]float64, numBands)
	maxs := make([]float64, numBands)
	for i := range mins {
		mins[i] = 1e16
	}
	for _, rr := range data {
		for i, v := range rr.Bands[:numBands] {
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
