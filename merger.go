package landsat

import (
	"fmt"
	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"math"
	"path"
	"sort"
	"strconv"
)

func MergeAndCorrect(root, prefix string, clip bool, options ...string) error {
	var (
		err error
		fo  = path.Join(root, prefix+".tiff")
		w   dataset.MultiBandWriter
		im  ImageMetadata
		buf []float64
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
			count, dist := distro(buf)
			if band != 6 && clip {
				// clip outliers outside of [0.1, 99.9] percentiles
				for i, v := range buf {
					if v < dist[0] || v > dist[4] {
						buf[i] = math.NaN()
					}
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
				Metadata("RADIATION_MAX", format(im.Bands[band].RadMax)).
				Metadata("GOOD_VALUES", strconv.Itoa(count)).
				Metadata("QUANTILES", fmt.Sprintf("0.1%%: %f, 25%%: %f, 50%%: %f, 75%%: %f, 99.9%%: %f", dist[0], dist[1], dist[2], dist[3], dist[4]))
			if band != 6 {
				rpb = rpb.Metadata("CORRECTION_FORMULA", fmt.Sprintf("(%fx + %f)/%f", scale, offset, div))
				if clip {
					rpb = rpb.Metadata("CLIPPED", "TRUE")
				}
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

func distro(buf []float64) (count int, dist [5]float64) {
	var data []float64
	for _, v := range buf {
		if !math.IsNaN(v) {
			data = append(data, v)
		}
	}
	count = len(data)
	sort.Float64s(data)
	dist[0] = data[count/1000-1]
	dist[1] = data[count/4-1]
	dist[2] = data[count/2-1]
	dist[3] = data[(count/4)*3-1]
	dist[4] = data[(count/1000)*999-1]
	return count, dist
}
