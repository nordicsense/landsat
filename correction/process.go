package correction

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
)

func MergeAndApply(pathIn, prefix string, pathOut string, options ...string) error {
	var (
		err error
		fo  = path.Join(pathOut, prefix+".tiff")
		w   dataset.MultiBandWriter
		im  dataset.ImageMetadata
		buf []float64
	)

	if im, err = dataset.ParseMetadata(pathIn, prefix); err != nil {
		return err
	}

	for band := 1; band <= 7; band++ {
		fi := path.Join(pathIn, prefix+"_B"+strconv.Itoa(band)+".TIF")
		if band == 6 {
			if _, err := os.Stat(fi); errors.Is(err, os.ErrNotExist) {
				fi = path.Join(pathIn, prefix+"_B6_VCID_1.TIF")
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
			// apply correction
			scale := im.Bands[band].RefScale
			offset := im.Bands[band].RefOffset
			div := math.Sin(im.SunElevation * math.Pi / 180.)
			if band != 6 {
				correct := func(v float64) float64 {
					return (scale*v + offset) / div
				}
				for i, v := range buf {
					buf[i] = correct(v)
				}
			}
			rpb := r.RasterParams().ToBuilder().Scale(1.0).Offset(0.0).
				Metadata("REFLECTION_SCALE", format(im.Bands[band].RefScale)).
				Metadata("REFLECTION_OFFSET", format(im.Bands[band].RefOffset)).
				Metadata("REFLECTION_MIN", format(im.Bands[band].RefMin)).
				Metadata("REFLECTION_MAX", format(im.Bands[band].RefMax)).
				Metadata("RADIATION_SCALE", format(im.Bands[band].RadScale)).
				Metadata("RADIATION_OFFSET", format(im.Bands[band].RadOffset)).
				Metadata("RADIATION_MIN", format(im.Bands[band].RadMin)).
				Metadata("RADIATION_MAX", format(im.Bands[band].RadMax))
			//				Metadata("QUANTILES", fmt.Sprintf("1%%: %f, 25%%: %f, 50%%: %f, 75%%: %f, 99%%: %f", dist[0], dist[1], dist[2], dist[3], dist[4]))
			if band != 6 {
				rpb = rpb.Metadata("CORRECTION_FORMULA", fmt.Sprintf("(%fx + %f)/%f", scale, offset, div))
			}
			bw := w.Writer(band)
			if err = bw.SetRasterParams(rpb.Build()); err == nil {
				err = bw.WriteBlock(0, 0, box, buf)
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
