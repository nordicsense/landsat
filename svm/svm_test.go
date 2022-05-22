package svm_test

import (
	"fmt"
	"log"
	"math"
	"testing"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"github.com/nordicsense/landsat/field"
	"github.com/nordicsense/landsat/svm"

	libSvm "github.com/nordicsense/libsvm-go"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func TestProcess(t *testing.T) {
	fieldDataPathIn := "/Users/osklyar/Data/Landsat/TrainingSet"
	imgPathIn := "/Users/osklyar/Data/Landsat/analysis/training"
	coord, err := field.Coordinates(fieldDataPathIn)
	if err != nil {
		t.Fatal(err)
	}
	data, err := field.TrainingDataOld(imgPathIn, ".*_T1.tiff", coord)
	if err != nil {
		t.Fatal(err)
	}

	costmax, gammamax, accmax, model := svm.Process(data)
	fmt.Printf("FINAL: %v,%v,%v\n", costmax, gammamax, accmax)
	err = model.Dump("/Users/osklyar/Data/Landsat/analysis/model3")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClassDistribution(t *testing.T) {
	fieldDataPathIn := "/Users/osklyar/Data/Landsat/TrainingSet"
	imgPathIn := "/Users/osklyar/Data/Landsat/analysis/training"
	coord, err := field.Coordinates(fieldDataPathIn)
	if err != nil {
		t.Fatal(err)
	}
	_, err = field.TrainingDataOld(imgPathIn, ".*_T1.tiff", coord)
	if err != nil {
		t.Fatal(err)
	}

}

func TestPrediction(t *testing.T) {
	model, err := libSvm.NewModelFromFile("/Users/osklyar/Data/Landsat/analysis/model2")
	if err != nil {
		t.Fatal(err)
	}
	mins := []float64{0.08045707311895159, 0.047084756609466344, 0.028087198320362303, 0.01873247532380952, 0.004349276646583651, 96, 0.002168673662203623}
	maxs := []float64{0.5036383271217346, 0.5171300570170084, 0.5264129704899259, 0.524611665142907, 0.41439854105313617, 143.55555555555554, 0.27334510617785984}
	normalize := svm.PixelToSVNormalizer(mins, maxs)

	r, err := dataset.OpenMultiBand("/Users/osklyar/Data/Landsat/analysis/prod/LT05_L1TP_188012_19900723_20200915_02_T1.tiff")
	if err != nil {
		t.Fatal(err)
	}

	// ugly performance workaround
	ds := r.Reader(1).BreakGlass()

	ip := r.ImageParams().ToBuilder().DataType(gdal.Byte).NaN(0.).Build()
	rp := r.Reader(1).RasterParams().ToBuilder().Offset(0.).Scale(1.).Build()

	w, err := dataset.NewUniBand("/Users/osklyar/Data/Landsat/analysis/model3-3x3-LT05_L1TP_188012_19900723_20200915_02_T1.tiff", dataset.GTiff, ip, rp)
	if err != nil {
		t.Fatal(err)
	}

	wds := w.BreakGlass().RasterBand(1)

	//nx := r.ImageParams().XSize()
	ny := r.ImageParams().YSize()

	x0 := 4000 // 0
	dx := 2000 // nx

	rrs := make([][]float64, 7)
	rr := make([]float64, 7)

	row1 := make([]float64, dx+2)
	row2 := make([]float64, dx+2)
	row3 := make([]float64, dx+2)
	firstread := true

	for j := 4000; j < 6000 && j < ny; j++ {
		fmt.Printf("Working on scan %d of %d", j-x0, dx)
		for band := 0; band < 7; band++ {
			if firstread {
				if err = ds.RasterBand(band+1).IO(gdal.Read, x0-1, j-1, dx+2, 1, row1, dx+2, 1, 0, 0); err != nil {
					t.Fatal(err)
				}
				if err = ds.RasterBand(band+1).IO(gdal.Read, x0-1, j, dx+2, 1, row2, dx+2, 1, 0, 0); err != nil {
					t.Fatal(err)
				}
				firstread = false
			} else {
				copy(row1, row2)
				copy(row2, row3)
			}
			if err = ds.RasterBand(band+1).IO(gdal.Read, x0-1, j+1, dx+2, 1, row3, dx+2, 1, 0, 0); err != nil {
				t.Fatal(err)
			}
			rrs[band] = make([]float64, dx)
			for i := 1; i <= dx; i++ {
				rrs[band][i-1] = (row1[i-1] + row1[i] + row1[i+1] + row2[i-1] + row2[i] + row2[i+1] + row3[i-1] + row3[i] + row3[i+1]) / 9.
			}
		}
		row := make([]int8, dx)
		stats := make(map[int8]int)
		for i := 0; i < dx; i++ {
			shouldSkip := false
			for band := 0; band < 7 && !shouldSkip; band++ {
				val := rrs[band][i]
				if band < 4 && math.IsNaN(val) {
					shouldSkip = true
				}
				rr[band] = val
			}
			var res int8 = -1
			if !shouldSkip {
				sv := libSvm.NewDenseSV(0., normalize(rr)...)
				res = int8(model.PredictVector(sv.Nodes))
			}
			row[i] = res
			stats[res]++
		}
		err = wds.IO(gdal.Write, x0, j, dx, 1, row, dx, 1, 0, 0)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf(": %v\n", stats)
	}
	w.Close()
}

func TestPrediction1(t *testing.T) {
	model, err := libSvm.NewModelFromFile("/Users/osklyar/Data/Landsat/analysis/model2")
	if err != nil {
		t.Fatal(err)
	}
	mins := []float64{0.08045707311895159, 0.047084756609466344, 0.028087198320362303, 0.01873247532380952, 0.004349276646583651, 96, 0.002168673662203623}
	maxs := []float64{0.5036383271217346, 0.5171300570170084, 0.5264129704899259, 0.524611665142907, 0.41439854105313617, 143.55555555555554, 0.27334510617785984}
	normalize := svm.PixelToSVNormalizer(mins, maxs)

	r, err := dataset.OpenMultiBand("/Users/osklyar/Data/Landsat/analysis/prod/LT05_L1TP_188012_19900723_20200915_02_T1.tiff")
	if err != nil {
		t.Fatal(err)
	}

	// ugly performance workaround
	ds := r.Reader(1).BreakGlass()

	ip := r.ImageParams().ToBuilder().DataType(gdal.Byte).NaN(0.).Build()
	rp := r.Reader(1).RasterParams().ToBuilder().Offset(0.).Scale(1.).Build()

	w, err := dataset.NewUniBand("/Users/osklyar/Data/Landsat/analysis/model3-3x3-LT05_L1TP_188012_19900723_20200915_02_T1.tiff", dataset.GTiff, ip, rp)
	if err != nil {
		t.Fatal(err)
	}

	wds := w.BreakGlass().RasterBand(1)

	//nx := r.ImageParams().XSize()
	ny := r.ImageParams().YSize()

	x0 := 2000 // 0
	dx := 4000 // nx

	rrs := make([][]float64, 7)
	rr := make([]float64, 7)

	for j := 2000; j < 6000 && j < ny; j++ {
		fmt.Printf("Working on scan %d of %d", j-x0, dx)
		for band := 0; band < 7; band++ {
			rrs[band] = make([]float64, dx)
			if err = ds.RasterBand(band+1).IO(gdal.Read, x0, j, dx, 1, rrs[band], dx, 1, 0, 0); err != nil {
				t.Fatal(err)
			}
		}
		row := make([]int8, dx)
		stats := make(map[int8]int)
		for i := 0; i < dx; i++ {
			shouldSkip := false
			for band := 0; band < 7 && !shouldSkip; band++ {
				val := rrs[band][i]
				if band < 4 && math.IsNaN(val) {
					shouldSkip = true
				}
				rr[band] = val
			}
			var res int8 = -1
			if !shouldSkip {
				sv := libSvm.NewDenseSV(0., normalize(rr)...)
				res = int8(model.PredictVector(sv.Nodes))
			}
			row[i] = res
			stats[res]++
		}
		err = wds.IO(gdal.Write, x0, j, dx, 1, row, dx, 1, 0, 0)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf(": %v\n", stats)
	}
	w.Close()
}
