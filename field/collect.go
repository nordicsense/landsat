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
	Clazz, Subclazz string
	Source          string
	Coords          [2]int
	Bands           []float64
}

var Mapping = map[string][2]string{
	"cloud":                          {"white", "cloud"},
	"snow":                           {"white", "snow"},
	"quarry":                         {"infra", "quarry"},
	"spoil_heap":                     {"infra", "spoil-heap"},
	"residential_area":               {"infra", "residential"},
	"road":                           {"infra", "road"},
	"asphalt":                        {"infra", "road"},
	"industrial_area":                {"infra", "industrial"},
	"tundra_stone_tundra":            {"rock", "stone"},
	"tundra_undam_stone_with_lichen": {"rock", "tundra"},
	"stone_dry_river_in_mountain":    {"rock", "dry-waterbed"},
	"dry_tailing_pond":               {"damaged", "dry-waterbed"},
	"human_forest_technogenic_barren_with_no_vegetation":           {"damaged", "barren"},
	"human_technogenic_barren_almost_with_no_vegetation":           {"damaged", "barren"},
	"human_severely_damaged":                                       {"damaged", "barren"},
	"new_burnt_area":                                               {"fire", "new"},
	"old_burnt_area":                                               {"fire", "old"},
	"natural_undam_birch_forest_with_dwarf_shrub_lichen":           {"birch", "shrubs"},
	"natural_undam_birch_forest_with_grass":                        {"birch", "grass"},
	"natural_undam_birch_forest_with_lichen_dwarf_shrub":           {"birch", "lichen"},
	"natural_undam_birch_spruce_forest_with_moss_lichen":           {"birch", "spruce"},
	"human_mostly_damaged_birch_spruce":                            {"birch", "damaged"}, // damaged?
	"natural_undam_pine_forest_with_dwarf_shrub":                   {"pine", "pine"},
	"natural_undam_pine_forest_with_dwarf_shrub_and_lichen":        {"pine", "lichen"},
	"natural_undam_pine_forest_with_dwarf_shrub_and_moss-lichen":   {"pine", "moss"},
	"natural_undam_pine_spruce_forest_with_dwarf_shrub":            {"pine", "spruce"},
	"natural_undam_spruce_forest_with_dwarf_shrub":                 {"spruce", "shrubs"},
	"natural_undam_spruce_forest_with_dwarf_shrub_and_moss-lichen": {"spruce", "moss"},
	"human_moderately_damaged_spruce_forest":                       {"spruce", "damaged"},
	"tundra_undam_lichen":                                          {"tundra", "lichen"},
	"tundra_undam_lichen_dwarf_shrub":                              {"tundra", "shrubs"},
	"water_with_no_sediments":                                      {"water", "clean"},
	"water_with_sediments":                                         {"polluted", "sediments"},
	"industrial_water":                                             {"polluted", "industrial"},
	"very_wet_tailing_pond":                                        {"polluted", "industrial"},
	"wet_tailing_pond":                                             {"polluted", "industrial"},
	"wetland_turf":                                                 {"wetland", "turf"},
	"wetland_with_dwarf_shrub_and_open_water":                      {"wetland", "water"},
	"wetland_with_dwarf_shrub_grass":                               {"wetland", "shrubs"},
	"wetland_with_dwarf_shrub_moss_grass":                          {"wetland", "moss"},
	"wetland_with_grass_moss_dwarf_shrub":                          {"wetland", "grass"},
	"agricultural_field_grass_birch_willow":                        {"grass", "birch"},
	"natural_undam_grey_willow_with_dwarf_shrub_grass":             {"grass", "willow"},
}

func TrainingData(pathIn, pattern string, coords map[string]coordinateMap) ([]Record, error) {
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
					rec := Record{
						Clazz:    Mapping[clazz][0],
						Subclazz: Mapping[clazz][1],
						Source:   im,
						Coords:   cc,
						Bands:    make([]float64, 7),
					}
					for band := 0; band < 7; band++ {
						// FIXME: coordinates are likely to be 1-based, thus -1
						if err = ds.RasterBand(band+1).IO(gdal.Read, cc[0]-1, cc[1]-1, 1, 1, buf, 1, 1, 0, 0); err != nil {
							return err
						}
						rec.Bands[band] = buf[0]
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
	return res, nil
}
