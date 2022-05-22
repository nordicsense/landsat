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

type Record struct {
	Image  string
	Clazz  string
	Coords [2]int
	Data   []float64
}

var NewMapping = map[string]string{

	"human_forest_technogenic_barren_with_no_vegetation": "I.1",
	"human_severely_damaged":                             "I.1",
	"human_technogenic_barren_almost_with_no_vegetation": "I.1",
	"old_burnt_area":        "I.5",
	"quarry":                "I.8",
	"industrial_area":       "I.8",
	"residential_area":      "I.8",
	"asphalt":               "I.8",
	"road":                  "I.8",
	"dry_tailing_pond":      "I.9",
	"very_wet_tailing_pond": "I.9",
	"water_with_sediments":  "I.9",
	"wet_tailing_pond":      "I.9",
	"natural_undam_pine_forest_with_dwarf_shrub_and_moss-lichen":   "II.1",
	"natural_undam_pine_forest_with_dwarf_shrub_and_lichen":        "II.1",
	"natural_undam_spruce_forest_with_dwarf_shrub":                 "II.2",
	"natural_undam_spruce_forest_with_dwarf_shrub_and_moss-lichen": "II.2",
	"natural_undam_birch_forest_with_grass":                        "II.3",
	"natural_undam_birch_forest_with_dwarf_shrub_lichen":           "II.3",
	"natural_undam_grey_willow_with_dwarf_shrub_grass":             "II.7",
	"wetland_with_dwarf_shrub_moss_grass":                          "II.8",
	"wetland_with_dwarf_shrub_and_open_water":                      "II.8",
	"tundra_undam_lichen":                                          "III.1",
	"tundra_undam_stone_with_lichen":                               "III.1",
	"cloud":                                                        "IV.1",
	"water_with_no_sediments":                                      "IV.3",
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

func TrainingDataOld(pathIn, pattern string, coords map[string]coordinateMap) ([]Record, error) {
	images := make(map[string]bool)
	for _, cm := range coords {
		for im := range cm {
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

	var res []Record

	stats1 := make(map[string]int)
	stats2 := make(map[string]int)

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

			// buf := make([]float64, 9)
			buf := make([]float64, 1)
			for clazz, cm := range coords {
				ccs, ok := cm[im]
				if !ok {
					continue
				}
				// ugly performance workaround
				ds := r.Reader(1).BreakGlass()
				for _, cc := range ccs {
					if _, ok := NewMapping[clazz]; !ok {
						continue
					}
					rec := Record{
						Clazz: NewMapping[clazz],
						//Subclazz: Mapping[clazz][1],
						Image:  im,
						Coords: cc,
						Data:   make([]float64, 7),
					}
					stats1[rec.Clazz]++
					for band := 0; band < 7; band++ {
						// FIXME: coordinates are likely to be 1-based, thus -1
						/*
							if err = ds.RasterBand(band+1).IO(gdal.Read, cc[0]-2, cc[1]-2, 3, 3, buf, 3, 3, 0, 0); err != nil {
								return err
							}
							mean := 0.
							for _, v := range buf {
								mean += v
							}
							mean /= 9.
							rec.Data[band] = mean
						*/
						if err = ds.RasterBand(band+1).IO(gdal.Read, cc[0]-1, cc[1]-1, 1, 1, buf, 1, 1, 0, 0); err != nil {
							return err
						}
						rec.Data[band] = buf[0]
					}
					res = append(res, rec)
				}
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	fmt.Println(stats1)
	fmt.Println(stats2)
	return res, nil
}
