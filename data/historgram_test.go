package data_test

import (
	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"math"
	"path"
	"testing"
)

func TestHistogram(t *testing.T) {
	root := "/Volumes/Caffeine/Data/Landsat/corrected"
	landsatId := 8
	r, err := dataset.OpenMultiBand(path.Join(root, "prod/LC08_L1TP_187012_20170710_20200903_02_T1.tiff"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	ds := r.Reader(1).BreakGlass()

	rrs := make([]float64, 4000)
	for band := 1; band <= 7; band++ {
		if landsatId == 8 && band == 1 {
			continue
		}
		if landsatId != 8 && band == 6 {
			continue
		}
		var buckets [100]int
		for y := 2000; y < 6000; y++ {
			if err = ds.RasterBand(band).IO(gdal.Read, 2000, y, 4000, 1, rrs, 4000, 1, 0, 0); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 4000; i++ {
				if !math.IsNaN(rrs[i]) {
					xx := rrs[i]
					/*
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
					*/
					ind := int((xx + 0.1) * 100)
					if ind < 0 || ind >= 100 {
						continue
					}
					buckets[ind]++
				}
			}
		}

		t.Log(buckets)
	}
}
