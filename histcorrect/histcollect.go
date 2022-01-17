package main

import (
	"encoding/csv"
	"math"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/nordicsense/landsat/dataset"
)

func HistCollect(root, prefix string, options ...string) error {
	var (
		err error
		r   dataset.MultiBandReader
		cf  *os.File
		w   *csv.Writer
		buf []float64
	)

	if r, err = dataset.OpenMultiBand(path.Join(root, prefix+".tiff")); err == nil {
		defer r.Close()
		if cf, err = os.Create(path.Join(root, prefix+"_hist.csv")); err == nil {
			w = csv.NewWriter(cf)
			defer func() {
				w.Flush()
				_ = cf.Close()
			}()
		}
	}
	if err != nil {
		return err
	}

	rs := []string{"", "", "", ""}
	for i := 0; err == nil && i < r.Bands(); i++ {
		br := r.Reader(i + 1)
		box := dataset.Box{0, 0, br.ImageParams().XSize(), br.ImageParams().YSize()}
		if buf, err = br.ReadBlock(0, 0, box); err == nil {
			min, max, hist := histogram(buf)
			delta := (max - min) / 100.

			rs[0] = strconv.Itoa(i + 1)
			for j := 0; err == nil && j < 100; j++ {
				if hist[j] == 0 {
					continue
				}
				rs[1] = strconv.Itoa(j + 1)
				rs[2] = strconv.FormatFloat(min+delta*float64(j+1), 'f', 6, 64)
				rs[3] = strconv.Itoa(hist[j])
				err = w.Write(rs)
			}
		}
	}
	return err
}

func histogram(buf []float64) (min, max float64, hist [100]int) {
	var data []float64
	for _, v := range buf {
		if !math.IsNaN(v) {
			data = append(data, v)
		}
	}
	sort.Float64s(data)

	min = data[0]
	max = data[len(data)-1]
	delta := (max - min) / 100.
	for i, j := 0, 0; i < 100; i++ {
		for ; j < len(data) && data[j] < min+delta*float64(i+1); j++ {
			hist[i] += 1
		}
	}
	return min, max, hist
}
