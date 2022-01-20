package field

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/nordicsense/gdal"
	"github.com/nordicsense/landsat/dataset"
	"github.com/nordicsense/landsat/io"
)

var coordRe = regexp.MustCompile(`^\s+(\d{1,4})\s+(\d{1,4})(?:\s+\d{1,3})+$`)

type coordinates [][2]int
type coordinateMap map[string]coordinates

// class -> image -> coodinates
func Coordinates(pathIn string) (map[string]coordinateMap, error) {
	var (
		err    error
		fNames []string
	)
	if fNames, err = io.ScanTree(pathIn, ".*.asc$"); err != nil {
		return nil, err
	}

	res := make(map[string]coordinateMap)
	for _, fName := range fNames {
		var (
			f   *os.File
			cm  coordinateMap
			ccs coordinates
			ok  bool
		)
		clazz := strings.Replace(path.Base(fName), ".asc", "", 1)
		if cm, ok = res[clazz]; !ok {
			cm = make(coordinateMap)
		}
		image := func() string {
			parts := strings.Split(path.Dir(fName), "/")
			return parts[len(parts)-3]
		}()
		ccs = cm[image]
		if f, err = os.Open(fName); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if coord := coordRe.FindAllStringSubmatch(line, 1); len(coord) == 1 && len(coord[0]) == 3 {
					// regex should guarantee conformance to int
					x, _ := strconv.Atoi(coord[0][1])
					y, _ := strconv.Atoi(coord[0][2])
					ccs = append(ccs, [2]int{x, y})
				}
			}
			err = scanner.Err()
			_ = f.Close()
		}
		if err != nil {
			break
		}
		// make conditional on non-empty content
		cm[image] = ccs
		res[clazz] = cm
	}
	return res, err
}

type Record []float64
type Records []Record

func TrainingData(pathIn, pattern string, coords map[string]coordinateMap) (map[string]Records, error) {
	images := make(map[string]bool)
	for _, cm := range coords {
		for im, _ := range cm {
			images[im] = true
		}
	}

	var (
		err    error
		fNames []string
	)
	if fNames, err = io.ScanTree(pathIn, pattern); err != nil {
		return nil, err
	}

	res := make(map[string]Records)

	for im := range images {
		log.Printf("collecting data from %s", im)
		fName := ""
		for _, aName := range fNames {
			if strings.Contains(aName, im) {
				fName = aName
				break
			}
		}
		if fName == "" {
			return nil, fmt.Errorf("could not find image for pattern %s", im)
		}

		err = func() error { // for scoping reader closure
			r, err := dataset.OpenMultiBand(fName)
			if err != nil {
				return err
			}
			defer r.Close()

			buf := make([]float64, 1)
			for clazz, cm := range coords {
				ccs, ok := cm[im]
				if !ok {
					continue
				}
				// ugly performance workaround
				ds := r.Reader(1).BreakGlass()
				for _, cc := range ccs {
					rec := make(Record, 7)
					for band := 0; band < 7; band++ {
						// FIXME: coordinates are likely to be 1-based, thus -1
						if err = ds.RasterBand(band+1).IO(gdal.Read, cc[0]-1, cc[1]-1, 1, 1, buf, 1, 1, 0, 0); err != nil {
							return err
						}
						rec[band] = buf[0]
					}
					res[clazz] = append(res[clazz], rec)
				}
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
