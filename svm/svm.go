package svm

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/nordicsense/landsat/field"
	libSvm "github.com/nordicsense/libsvm-go"
)

var clazzes = map[string]float64{
	"I.1":   1.,
	"I.2":   2.,
	"I.3":   3.,
	"I.4":   4.,
	"I.5":   5.,
	"I.7":   6.,
	"I.9":   7.,
	"II.1":  8.,
	"II.3":  9.,
	"II.7":  10.,
	"II.8":  11.,
	"III.1": 12.,
	"III.3": 13.,
	"IV.1":  14.,
	"IV.3":  15.,
}

func Process(data []field.Record) {
	svs, _ := toSVs(data, 1000)
	p, err := libSvm.NewProblem(svs)
	if err != nil {
		log.Fatal(err)
	}

	clss := make(map[int]string)
	for name, index := range clazzes {
		clss[int(index)] = name
	}

	// https://scikit-learn.org/stable/auto_examples/svm/plot_rbf_parameters.html
	par := &libSvm.Parameter{
		SvmType:    libSvm.C_SVC,
		KernelType: libSvm.RBF,
		Gamma:      0.2, //1. / float64(numBands),
		Eps:        1e-5,
		C:          750,
		//		NrWeight:    15,
		//		WeightLabel: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		//		Weight:      []float64{0.1, 0.2, 0.8, 0.2, 1.2, 0.3, 0.05, 2.2, 2.0, 1.0, 1.5, 0.2, 0.05, 0.05, 0.1},
		CacheSize: 2000,
		NumCPU:    4,
		QuietMode: true,
	}
	m := libSvm.NewModel(par)

	log.Printf("training model with %d vectors\n", len(svs))
	err = m.Train(p)
	if err != nil {
		log.Fatal(err)
	}
	var matches, mismatches int
	log.Printf("validating %d cases\n", len(svs))

	expected := make(map[string]int)
	found := make(map[string]int)
	pairs := make(map[string]int)

	for _, sv := range svs {
		lbl := m.PredictVector(sv.Nodes)
		ec := clss[int(sv.Label)]
		fc := clss[int(lbl)]
		expected[ec]++
		found[fc]++
		pair := fmt.Sprintf("%s-%s", ec, fc)
		pairs[pair]++
		if ec == fc {
			matches++
		} else {
			mismatches++
		}
	}
	log.Printf("accuracy %f\n", float64(matches)/float64(matches+mismatches))
	log.Println(expected)
	log.Println(found)
	log.Println(pairs)
}

func toSVs(rrs []field.Record, nSVs int) ([]libSvm.SV, map[string]float64) {
	norm := normalizer(rrs)

	xx := make(map[string][][]float64)
	for _, rr := range rrs {
		xx[rr.Clazz] = append(xx[rr.Clazz], rr.Bands)
	}

	var res []libSvm.SV
	for clazz, x := range xx {
		smpl := subsample(x, nSVs)
		for _, rr := range smpl {
			nrr := norm(rr)
			res = append(res, libSvm.NewDenseSV(clazzes[clazz], nrr...))
		}
	}
	return res, clazzes
}

func subsample(rrs [][]float64, nSVs int) [][]float64 {
	if nSVs >= len(rrs) {
		return rrs
	}
	idx := rand.Perm(len(rrs))
	res := make([][]float64, nSVs)
	for i := 0; i < nSVs; i++ {
		res[i] = rrs[idx[i]]
	}
	return res
}

func normalizer(data []field.Record) func([]float64) []float64 {
	bands := make(map[int][]float64)
	for _, rr := range data {
		for i, v := range rr.Bands {
			if math.IsNaN(v) {
				continue
			}
			band := append(bands[i], v)
			bands[i] = band
		}
	}

	mins := make([]float64, 7)
	maxs := make([]float64, 7)
	for i, band := range bands {
		sort.Float64s(band)
		mins[i] = band[len(band)/100*5]
		maxs[i] = band[len(band)/100*95]
	}
	log.Println(mins)
	log.Println(maxs)
	return func(xx []float64) []float64 {
		res := make([]float64, 9)
		res[0] = (xx[0]-mins[0])/(maxs[0]-mins[0]) - 0.5
		res[1] = (xx[1]-mins[1])/(maxs[1]-mins[1]) - 0.5
		res[2] = (xx[2]-mins[2])/(maxs[2]-mins[2]) - 0.5
		res[3] = (xx[3]-mins[3])/(maxs[3]-mins[3]) - 0.5
		res[4] = (xx[4]-mins[4])/(maxs[4]-mins[4]) - 0.5
		res[5] = (xx[6]-mins[6])/(maxs[6]-mins[6]) - 0.5
		res[6] = (xx[3] - xx[2]) / (xx[3] + xx[2]) // NDVI
		res[7] = (xx[3] - xx[6]) / (xx[3] + xx[6]) // NBR
		res[8] = (xx[3] - xx[4]) / (xx[3] + xx[4]) // NDWI
		return res
	}
}
