package data_test

import (
	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"math"
	"path"
	"sort"
	"testing"
)

func TestHistogram(t *testing.T) {
	root := "/Volumes/Caffeine/Data/Landsat/corrected/prod"
	landsatId := 5
	r, err := dataset.OpenMultiBand(path.Join(root, "LT05_L1TP_187013_20050709_20200902_02_T1.tiff"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	ds := r.Reader(1).BreakGlass()

	ip := r.ImageParams()
	nx := ip.XSize()
	ny := ip.YSize()

	rrs := make([]float64, nx)
	for band := 5; band <= 6; band++ {
		if landsatId == 8 && band == 1 {
			continue
		}
		if landsatId != 8 && band == 6 {
			continue
		}

		// detect approx range
		if err = ds.RasterBand(band).IO(gdal.Read, 0, 4000, nx, 1, rrs, nx, 1, 0, 0); err != nil {
			t.Fatal(err)
		}
		sort.Float64s(rrs)
		count := 0
		val := 0.
		for i := 4000; i < 4100; i++ {
			if !math.IsNaN(rrs[i]) {
				val += rrs[i]
				count++
			}
		}
		var buckets [102]int
		for y := 0; y < ny; y++ {
			if err = ds.RasterBand(band).IO(gdal.Read, 0, y, nx, 1, rrs, nx, 1, 0, 0); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < nx; i++ {
				if !math.IsNaN(rrs[i]) {
					xx := rrs[i]
					if band == 6 {
						xx = rrs[i] / 65535.
					}

					if landsatId == 8 {
						if band == 2 {
							xx = (xx-0.087)*1.4 + 0.087
						} else if band == 4 {
							xx = (xx-0.044)*1.2 + 0.044
						} else if band == 5 {
							xx = (xx-0.215)*0.95 + 0.215
						} else if band == 6 {
							xx = (xx-0.14)*2. + 0.14
						}
					} else if landsatId == 7 {
						if band == 3 {
							xx = (xx-0.049)*0.9 + 0.049
						}
					} else {
						if band == 5 {
							xx = (xx+0.005)*1.2 - 0.005
						}
					}

					ind := int(xx*100) + 1
					if ind < 0 {
						ind = 0
					} else if ind > 101 {
						ind = 101
					}
					buckets[ind]++
				}
			}
		}

		t.Log(buckets)
	}
}
