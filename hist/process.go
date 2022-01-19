package hist

import (
	"encoding/csv"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nordicsense/landsat/dataset"
)

func CollectForMergedImage(fName, pathOut string) error {
	var (
		err error
		r   dataset.MultiBandReader
		cf  *os.File
		w   *csv.Writer
		buf []float64
	)

	if r, err = dataset.OpenMultiBand(fName); err == nil {
		defer r.Close()
		fNameOut := path.Join(pathOut, strings.Replace(path.Base(fName), path.Ext(fName), "", 1)+"_hist.csv")
		if cf, err = os.Create(fNameOut); err == nil {
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
			min, max, hist := Compute(buf)
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
