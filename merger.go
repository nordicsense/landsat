package landsat

import (
	"fmt"
	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"log"
	"math"
	"path"
	"strconv"
)

func MergeAndCorrect(root, prefix string) error {
	var (
		err error
		fo  = path.Join(root, prefix+".tiff")
		w   dataset.MultiBandWriter
		im  ImageMetadata
	)

	if im, err = ParseMetadata(root, prefix); err != nil {
		return err
	}

	for band := 1; band <= 7; band++ {
		fi := path.Join(root, prefix+"_B"+strconv.Itoa(band)+".TIF")
		var r dataset.UniBandReader
		if r, err = dataset.OpenUniBand(fi); err != nil {
			if w != nil {
				w.Close()
			}
			return err
		}
		format := func(v float64) string {
			return strconv.FormatFloat(v, 'f', 6, 64)
		}

		// https://www.gisagmaps.com/landsat-8-atco/
		ip := r.ImageParams().ToBuilder().DataType(gdal.Float32).NaN(math.NaN()).Build()
		if w == nil {
			// best for float32, see https://kokoalberti.com/articles/geotiff-compression-optimization-guide/
			options := []string{"compress=deflate", "zlevel=6", "predictor=3"}
			if w, err = dataset.NewMultiBand(fo, dataset.GTiff, 7, ip, options...); err != nil {
				r.Close()
				break
			}
			// This hacks into the metadata of the multilayered image, which is not supported by dataset API
			ds := w.Writer(band).BreakGlass()
			// ignore errors setting these metadata
			_ = ds.SetMetadataItem("DATE", im.Date.Format("2006-01-02"), "")
			_ = ds.SetMetadataItem("SUN_ELEVATION", format(im.SunElevation), "")
			_ = ds.SetMetadataItem("SUN_AZIMUTH", format(im.SunAzimuth), "")
			for k, v := range im.Aux {
				_ = ds.SetMetadataItem(k, v, "")
			}
		}

		scale := im.Bands[band].RefScale
		offset := im.Bands[band].RefOffset
		div := math.Sin(im.SunElevation * math.Pi / 180.)
		correct := func(v float64) float64 {
			if band == 6 {
				return v
			}
			return (scale*v + offset) / div
		}

		rp := r.RasterParams().ToBuilder().Scale(1.0).Offset(0.0).
			Metadata("REFLECTION_SCALE", format(im.Bands[band].RefScale)).
			Metadata("REFLECTION_OFFSET", format(im.Bands[band].RefOffset)).
			Metadata("REFLECTION_MIN", format(im.Bands[band].RefMin)).
			Metadata("REFLECTION_MAX", format(im.Bands[band].RefMax)).
			Metadata("RADIATION_SCALE", format(im.Bands[band].RadScale)).
			Metadata("RADIATION_OFFSET", format(im.Bands[band].RadOffset)).
			Metadata("RADIATION_MIN", format(im.Bands[band].RadMin)).
			Metadata("RADIATION_MAX", format(im.Bands[band].RadMax)).
			Metadata("CORRECTION_FORMULA", fmt.Sprintf("(%fx + %f)/%f", scale, offset, div)).
			Build()
		bw := w.Writer(band)
		if err = bw.SetRasterParams(rp); err == nil {
			nx := ip.XSize()
			ny := ip.YSize()
			box := dataset.Box{0, 0, nx, ny}
			if buffer, err := r.ReadBlock(0, 0, box); err == nil {
				// apply correction
				for i, v := range buffer {
					buffer[i] = correct(v)
				}
				err = bw.WriteBlock(0, 0, box, buffer)

				// FIXME: drop
				var vals []float64
				for i := 0; i < 4; i++ {
					vals = append(vals, buffer[i+nx*4000+4000])
				}
				log.Println(vals)

			}
		}
		r.Close()
		if err != nil {
			break
		}
	}
	w.Close()
	return err
}

/*
L5:

2022/01/14 22:13:57 [61 59 56 58 56 56 57 56 57 58]
2022/01/14 22:13:58 [25 24 25 23 23 23 23 23 24 23]
2022/01/14 22:14:03 [22 20 19 17 16 16 16 17 18 19]
2022/01/14 22:14:29 [48 51 50 58 67 67 64 57 54 49]
2022/01/14 22:14:39 [61 59 57 53 52 50 52 53 56 54]
2022/01/14 22:15:00 [136 135 134 133 133 133 133 133 134 134]
2022/01/14 22:15:18 [22 24 21 18 16 17 18 20 21 20]

*/
