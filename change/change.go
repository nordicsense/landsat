package change

import (
	"fmt"
	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/classification"
	"github.com/nordicsense/landsat/dataset"
	"github.com/vardius/progress-go"
	"path"
)

var (
	TL = dataset.LatLon{7674459, 355738} // sin
	BR = dataset.LatLon{7339883, 593636}

	accepted_mismatch map[int]map[int]bool
)

func init() {
	accepted_mismatch = map[int]map[int]bool{
		2: {
			3: true,
		},
		3: {
			2: true,
		},
		4: {
			5: true,
		},
		5: {
			4: true,
		},
		6: {
			7:  true,
			10: true,
		},
		7: {
			6:  true,
			10: true,
		},
		10: {
			6: true,
			7: true,
		},
	}
}

func Collect(fromTiffs, toTiffs []string, tl, br dataset.LatLon, outputDir string) error {

	if len(fromTiffs) != 2 || len(toTiffs) != 2 {
		return fmt.Errorf("incorrect number of _from_ (%d) or _to_ (%d) images, expected 2 each", len(fromTiffs), len(toTiffs))
	}

	f0, err := dataset.OpenUniBand(fromTiffs[0])
	if err != nil {
		return err
	}
	defer f0.Close()
	f1, err := dataset.OpenUniBand(fromTiffs[1])
	if err != nil {
		return err
	}
	defer f1.Close()
	t0, err := dataset.OpenUniBand(toTiffs[0])
	if err != nil {
		return err
	}
	defer t0.Close()
	t1, err := dataset.OpenUniBand(toTiffs[1])
	if err != nil {
		return err
	}
	defer t1.Close()

	ip := f0.ImageParams().ToBuilder().Transform(dataset.AffineTransform{tl[1], 30., 0., tl[0], 0., -30.}).Build()

	tf := ip.Transform()
	ll, _ := br.Sin2Degree()
	nx, ny := tf.LatLon2Pixels(ll)

	ip = dataset.ImageParamsBuilder(nx, ny).Transform(tf).DataType(gdal.Int16).Projection(ip.Projection()).NaN(0).Build()

	rp := f0.RasterParams().ToBuilder().Build()

	w, err := dataset.NewUniBand(path.Join(outputDir, "out.tiff"), dataset.GTiff,
		ip, rp, "compress=LZW", "predictor=2")
	if err != nil {
		return err
	}
	defer w.Close()

	bar := progress.New(0, int64(ny))
	bar.Start()

	wds := w.BreakGlass().RasterBand(1)

	dsMap := make(map[dataset.UniBandReader]gdal.RasterBand)
	dsMap[f0] = f0.BreakGlass().RasterBand(1)
	dsMap[f1] = f1.BreakGlass().RasterBand(1)
	dsMap[t0] = t0.BreakGlass().RasterBand(1)
	dsMap[t1] = t1.BreakGlass().RasterBand(1)

	var m [classification.NClasses][classification.NClasses]int
	for y := 0; y < ny; y++ {
		ll := tf.Pixels2LatLon(0, y)

		row1 := make([]float64, nx)
		row2 := make([]float64, nx)
		var err error

		_, yyf := f0.ImageParams().Transform().LatLon2Pixels(ll)
		_, yyt := t0.ImageParams().Transform().LatLon2Pixels(ll)
		if yyf >= 0 && yyt >= 0 && yyf < f0.ImageParams().YSize() && yyt < t0.ImageParams().YSize() {
			x0f, _ := tf.LatLon2Pixels(f0.ImageParams().Transform().Pixels2LatLon(0, yyf))
			x0t, _ := tf.LatLon2Pixels(t0.ImageParams().Transform().Pixels2LatLon(0, yyt))
			row1, err = merge(ip, f0, t0, x0f, yyf, x0t, yyt, &m, dsMap)
			if err != nil {
				return err
			}
		}
		_, yyf = f1.ImageParams().Transform().LatLon2Pixels(ll)
		_, yyt = t1.ImageParams().Transform().LatLon2Pixels(ll)
		if yyf >= 0 && yyt >= 0 && yyf < f1.ImageParams().YSize() && yyt < t1.ImageParams().YSize() {
			x0f, _ := tf.LatLon2Pixels(f1.ImageParams().Transform().Pixels2LatLon(0, yyf))
			x0t, _ := tf.LatLon2Pixels(t1.ImageParams().Transform().Pixels2LatLon(0, yyt))
			row2, err = merge(ip, f1, t1, x0f, yyf, x0t, yyt, &m, dsMap)
			if err != nil {
				return err
			}
		}
		for i, v1 := range row1 {
			if v1 == 0 {
				row1[i] = row2[i]
			}
		}
		if err = wds.IO(gdal.Write, 0, y, nx, 1, row1, nx, 1, 0, 0); err != nil {
			return err
		}
		bar.Advance(1)
	}
	bar.Stop()

	fmt.Println(m)
	return nil
}

func merge(ip *dataset.ImageParams, f0 dataset.UniBandReader, t0 dataset.UniBandReader, x0f, yyf, x0t, yyt int, m *[classification.NClasses][classification.NClasses]int, dsMap map[dataset.UniBandReader]gdal.RasterBand) ([]float64, error) {
	row := make([]float64, f0.ImageParams().XSize())
	if err := dsMap[f0].IO(gdal.Read, 0, yyf, f0.ImageParams().XSize(), 1, row, f0.ImageParams().XSize(), 1, 0, 0); err != nil {
		return nil, err
	}
	fRow := make([]float64, ip.XSize())
	for i, v := range row {
		if v == 0. {
			continue
		}
		xx := i + x0f
		if xx >= 0 && xx < ip.XSize() {
			fRow[xx] = v
		}
	}
	row = make([]float64, t0.ImageParams().XSize())
	if err := dsMap[t0].IO(gdal.Read, 0, yyt, t0.ImageParams().XSize(), 1, row, t0.ImageParams().XSize(), 1, 0, 0); err != nil {
		return nil, err
	}
	tRow := make([]float64, ip.XSize())
	for i, v := range row {
		if v == 0. {
			continue
		}
		xx := i + x0t
		if xx >= 0 && xx < ip.XSize() {
			tRow[xx] = v
		}
	}

	row = make([]float64, ip.XSize())
	for i, fv := range fRow {
		tv := tRow[i]
		tvi := int(tv)
		fvi := int(fv)
		if fvi != 0 && tvi != 0 {
			if accepted_mismatch[fvi][tvi] {
				tvi = fvi
				tv = fv
			}
			if fvi == tvi {
				row[i] = fv
			} else {
				row[i] = -1.
			}
			m[fvi-1][tvi-1]++
		}
	}
	return row, nil
}
