package field

import (
	"bufio"
	"fmt"
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
func CollectCoordinates(pathIn string) (map[string]coordinateMap, error) {
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

type Record struct {
	Image  string
	Clazz  string
	Coords [2]int
	Data   []float64
}

type Converter func(string, string, [2]int, []float64) (string, []float64, bool)

var PathThrough Converter = func(image string, clazz string, coord [2]int, data []float64) (string, []float64, bool) {
	return clazz, data, true
}

func TrainingData(pathIn, pattern string, coords map[string]coordinateMap, convert Converter) ([]Record, error) {
	var (
		err         error
		imageFNames []string
		imageNames  = make(map[string]bool)
		res         []Record
	)
	if imageFNames, err = io.ScanTree(pathIn, pattern); err != nil {
		return nil, err
	}

	for _, cm := range coords {
		for im := range cm {
			imageNames[im] = true
		}
	}

	for im := range imageNames {
		fName := ""
		for _, n := range imageFNames {
			if strings.Contains(n, im) {
				fName = n
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
				// ugly performance workaround to get access to the raw reader
				ds := r.Reader(1).BreakGlass()
				for _, cc := range ccs {
					data := make([]float64, 7)
					for band := 0; band < 7; band++ {
						if err = ds.RasterBand(band+1).IO(gdal.Read, cc[0]-1, cc[1]-1, 1, 1, buf, 1, 1, 0, 0); err != nil {
							return err
						}
						data[band] = buf[0]
					}
					newclazz, newdata, ok := convert(im, clazz, cc, data)
					if ok {
						res = append(res, Record{
							Image:  im,
							Clazz:  newclazz,
							Coords: cc,
							Data:   newdata,
						})
					}
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
