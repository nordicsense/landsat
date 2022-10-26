package classification

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/nordicsense/landsat/data"
)

type classIdMap struct {
	clazz string
	index int
}

var (
	mapping = map[string]classIdMap{
		"cloud": {"cloud", 0},

		"water_with_no_sediments": {"water", 1},

		"wet_tailing_pond":      {"water-dam", 2},
		"water_with_sediments":  {"water-dam", 2},
		"very_wet_tailing_pond": {"water-dam", 2},
		"industrial_water":      {"water-dam", 2},

		"dry_tailing_pond": {"non-veg", 3},
		"residential_area": {"non-veg", 3},
		"asphalt":          {"non-veg", 3},
		"quarry":           {"non-veg", 3},
		"industrial_area":  {"non-veg", 3},
		"human_technogenic_barren_almost_with_no_vegetation": {"non-veg", 3},
		"road":                   {"non-veg", 3},
		"human_severely_damaged": {"non-veg", 3},
		"human_forest_technogenic_barren_with_no_vegetation": {"non-veg", 3},
		"spoil_heap":                  {"non-veg", 3},
		"stone_dry_river_in_mountain": {"non-veg", 3},
		"tundra_stone_tundra":         {"non-veg", 3},

		"new_burnt_area": {"burnt", 4},

		"wetland_with_dwarf_shrub_and_open_water":          {"dwarf-shrub", 5},
		"natural_undam_grey_willow_with_dwarf_shrub_grass": {"dwarf-shrub", 5},

		"wetland_with_dwarf_shrub_grass": {"wetland", 6},

		"natural_undam_pine_spruce_forest_with_dwarf_shrub":     {"pine", 7},
		"natural_undam_pine_forest_with_dwarf_shrub_and_lichen": {"pine", 7},
		"human_moderately_damaged_spruce_forest":                {"pine", 7},

		"natural_undam_spruce_forest_with_dwarf_shrub_and_moss-lichen": {"spruce", 8},
		"natural_undam_spruce_forest_with_dwarf_shrub":                 {"spruce", 8},
		"natural_undam_pine_forest_with_dwarf_shrub_and_moss-lichen":   {"spruce", 8},
		"natural_undam_pine_forest_with_dwarf_shrub":                   {"spruce", 8},

		"natural_undam_birch_forest_with_dwarf_shrub_lichen": {"deciduous", 9},
		"natural_undam_birch_forest_with_grass":              {"deciduous", 9},
		"natural_undam_birch_spruce_forest_with_moss_lichen": {"deciduous", 9},

		"tundra_undam_lichen_dwarf_shrub": {"veg-tundra", 10},
		"tundra_undam_lichen":             {"veg-tundra", 10},
		"tundra_undam_stone_with_lichen":  {"veg-tundra", 10},
	}

	images = map[string]bool{
		"LT05_L1TP_190011_20090725": true,
		"LT05_L1TP_187012_20050709": true,
		"LE07_L1TP_186012_20000728": true,
		"LT05_L1TP_188012_19860728": true,

		"LE07_L1TP_188012_20000726": true,
		"LT05_L1TP_190012_19930713": true,
	}

	r             *rand.Rand
	ClassNameToId map[string]int
	ClassIdToName map[int]string
)

const NClasses = 11

func init() {
	r = rand.New(rand.NewSource(42))
	ClassNameToId = make(map[string]int)
	ClassIdToName = make(map[int]string)
	for _, m := range mapping {
		ClassNameToId[m.clazz] = m.index
		ClassIdToName[m.index] = m.clazz
	}
	if NClasses != len(ClassNameToId) {
		panic(fmt.Sprintf("incorrect mapping, got %d classes, want %d", len(ClassNameToId), NClasses))
	}
}

const (
	trainFraction = 0.8
	clazzSize     = 40000
	testSize      = 3000
)

func CollectTrainingData(tabPath, imgPath, csvPath, imgPattern string) error {
	coord, err := data.CollectCoordinates(tabPath)
	if err != nil {
		return err
	}
	recs, err := data.TrainingData(imgPath, imgPattern, coord, images, convert)
	if err != nil {
		return err
	}
	train, test := data.Subsample(recs, ClassNameToId, clazzSize, testSize, r, trainFraction)
	if err = data.DumpCSV(csvPath+".csv", train, ClassNameToId); err != nil {
		return err
	}
	if err = data.DumpCSV(csvPath+"-test.csv", test, ClassNameToId); err != nil {
		return err
	}

	stats := make([]int, NClasses)
	for _, r := range train {
		j := ClassNameToId[r.Clazz]
		stats[j]++
	}
	log.Println("training stats:", stats)

	stats = make([]int, NClasses)
	for _, r := range test {
		j := ClassNameToId[r.Clazz]
		stats[j]++
	}
	log.Println("testing stats:", stats)
	return nil
}

func convert(im string, clazz string, xx []float64) (string, []float64, bool) {
	var ok bool
	newClazz, ok := mapping[clazz]
	if !ok {
		return "", nil, false
	}
	if !images[im] {
		return "", nil, false
	}
	for _, x := range xx {
		if math.IsNaN(x) {
			return "", nil, false
		}
	}
	landsatId := 7
	switch im[0:5] {
	case "LT05":
		landsatId = 5
	case "LC08":
		landsatId = 8
	}
	return newClazz.clazz, data.Transform(xx, landsatId), true
}
