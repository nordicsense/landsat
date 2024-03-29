package conversion

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"github.com/vardius/progress-go"
)

func MergeAndApply(pathIn, prefix string, pathOut string, l1, skip, verbose bool, options ...string) error {
	var (
		err error
		fo  = path.Join(pathOut, prefix+".tiff")
		w   dataset.MultiBandWriter
		im  dataset.ImageMetadata
		buf []float64
	)

	if _, err := os.Stat(fo); skip && err == nil {
		return nil
	}

	if im, err = dataset.ParseMetadata(pathIn, prefix); err != nil {
		return err
	}

	bar := progress.New(0, 7)
	if verbose {
		bar.Start()
	}

	for band := 1; band <= 7; band++ {
		fi := path.Join(pathIn, prefix+"_B"+strconv.Itoa(band)+".TIF")
		if band == 6 {
			if _, err := os.Stat(fi); errors.Is(err, os.ErrNotExist) {
				fi = path.Join(pathIn, strings.Replace(prefix, "SR", "ST", 1)+"_B6.TIF")
				if _, err := os.Stat(fi); errors.Is(err, os.ErrNotExist) {
					fi = path.Join(pathIn, prefix+"_B6_VCID_1.TIF")
				}
			}
		}
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

		box := dataset.Box{0, 0, ip.XSize(), ip.YSize()}
		if buf, err = r.ReadBlock(0, 0, box); err == nil {
			var dist [10]float64
			// apply scaling
			scale := 2.75e-05
			offset := -0.2
			div := 1.0
			if l1 {
				// Perform ToA radiance conversion
				scale = im.Bands[band].RefScale
				offset = im.Bands[band].RefOffset
				// generally useless as all the data seems to be missing at the same time
				if scale == 1.0 && offset == 0.0 && !math.IsNaN(im.Bands[band].RefMax) && !math.IsNaN(im.Bands[band].RefMin) {
					scale = (im.Bands[band].RefMax - im.Bands[band].RefMin) / 255.
					offset = im.Bands[band].RefMin
				}
				div = math.Sin(im.SunElevation * math.Pi / 180.)
			}
			correct := func(v float64) float64 {
				return (scale*v + offset) / div
			}
			for i, v := range buf {
				v = correct(v)
				buf[i] = v
				if !math.IsNaN(v) {
					j := int(v * 10)
					if j < 0 {
						j = 0
					} else if j > 9 {
						j = 9
					}
					dist[j]++
				}
			}
			rpb := r.RasterParams().ToBuilder().Scale(1.0).Offset(0.0)

			if l1 {
				rpb = rpb.
					Metadata("REFLECTION_SCALE", format(scale)).
					Metadata("REFLECTION_OFFSET", format(offset)).
					Metadata("REFLECTION_MIN", format(im.Bands[band].RefMin)).
					Metadata("REFLECTION_MAX", format(im.Bands[band].RefMax)).
					Metadata("RADIATION_SCALE", format(im.Bands[band].RadScale)).
					Metadata("RADIATION_OFFSET", format(im.Bands[band].RadOffset)).
					Metadata("RADIATION_MIN", format(im.Bands[band].RadMin)).
					Metadata("RADIATION_MAX", format(im.Bands[band].RadMax))
			}
			rpb = rpb.Metadata("DIST", fmt.Sprintf("%v", dist))
			rpb = rpb.Metadata("CORRECTION_FORMULA", fmt.Sprintf("(%fx + %f)/%f", scale, offset, div))
			bw := w.Writer(band)
			if err = bw.SetRasterParams(rpb.Build()); err == nil {
				err = bw.WriteBlock(0, 0, box, buf)
			}
		}
		r.Close()
		if err != nil {
			break
		}
		if verbose {
			bar.Advance(1)
		}
	}
	w.Close()
	bar.Stop()
	return err
}
