package training

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/nordicsense/landsat/data"
	"github.com/nordicsense/landsat/training/collector"
)

type classIdMap struct {
	clazz string
	index int
}

var (
	mapping = map[string]classIdMap{
		"snow":                    {"snow", 0},
		"cloud":                   {"cloud", 1},
		"water_with_no_sediments": {"water", 2},
		"wet_tailing_pond":        {"water-dam", 3},
		"water_with_sediments":    {"water-dam", 3},
		"very_wet_tailing_pond":   {"water-dam", 3},
		"industrial_water":        {"water-dam", 3},
		"dry_tailing_pond":        {"non-veg", 4},
		"residential_area":        {"non-veg", 4},
		"asphalt":                 {"non-veg", 4},
		"quarry":                  {"non-veg", 4},
		"industrial_area":         {"non-veg", 4},
		"human_technogenic_barren_almost_with_no_vegetation": {"non-veg", 4},
		"human_severely_damaged":                             {"non-veg", 4},
		"road":                                               {"non-veg", 4},
		"human_forest_technogenic_barren_with_no_vegetation": {"non-veg", 4},
		"spoil_heap":                  {"non-veg", 4},
		"stone_dry_river_in_mountain": {"non-veg", 4},
		"tundra_stone_tundra":         {"non-veg", 4},

		"agricultural_field_grass_birch_willow": {"agricultural_field_grass_birch_willow", 5},

		"natural_undam_grey_willow_with_dwarf_shrub_grass": {"wetland", 6},
		"wetland_with_dwarf_shrub_grass":                   {"wetland", 6},
		"wetland_with_dwarf_shrub_moss_grass":              {"wetland", 6},
		"wetland_with_grass_moss_dwarf_shrub":              {"wetland", 6}, // only: LE07_L1TP_195012_20000727

		"wetland_with_dwarf_shrub_and_open_water": {"wetland_open_water", 7},
		"wetland_turf": {"wetland_turf", 8},

		"human_moderately_damaged_spruce_forest": {"human_moderately_damaged_spruce_forest", 9},
		"human_mostly_damaged_birch_spruce":      {"human_mostly_damaged_birch_spruce", 10},

		"new_burnt_area": {"burnt", 11},
		"old_burnt_area": {"burnt", 11},

		"natural_undam_pine_forest_with_dwarf_shrub_and_lichen":      {"pine", 12},
		"natural_undam_pine_forest_with_dwarf_shrub_and_moss-lichen": {"pine", 12},
		"natural_undam_pine_forest_with_dwarf_shrub":                 {"pine", 12},

		"natural_undam_spruce_forest_with_dwarf_shrub_and_moss-lichen": {"spruce", 13},
		"natural_undam_spruce_forest_with_dwarf_shrub":                 {"spruce", 13},

		"natural_undam_birch_forest_with_dwarf_shrub_lichen": {"birch", 14},
		"natural_undam_birch_forest_with_grass":              {"birch", 14},

		"tundra_undam_lichen_dwarf_shrub": {"veg-tundra", 15},
		"tundra_undam_lichen":             {"veg-tundra", 15},
		"tundra_undam_stone_with_lichen":  {"veg-tundra", 15},
	}

	images = map[string]bool{
		"LT05_L1TP_190011_20090725": true,
		"LT05_L1TP_187012_20050709": true,
		"LE07_L1TP_186012_20000728": true,
		"LT05_L1TP_188012_19860728": true,

		"LE07_L1TP_188012_20000726": true,  // false
		"LE07_L1TP_195011_20000727": false, // false
		"LE07_L1TP_195012_20000727": true,  // false
		"LT05_L1TP_190012_19930713": true,  // false
	}

	r             *rand.Rand
	ClassNameToId map[string]int
	ClassIdToName map[int]string
)

const NClasses = 16

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
	trainFraction = 0.9
	clazzSize     = 50000
	testSize      = 4000
)

func Collect(tabPath, imgPath, csvPath, imgPattern string) error {
	coord, err := collector.CollectCoordinates(tabPath)
	if err != nil {
		return err
	}
	recs, err := collector.TrainingData(imgPath, imgPattern, coord, images, convert)
	if err != nil {
		return err
	}
	train, test := collector.Subsample(recs, ClassNameToId, clazzSize, testSize, r, trainFraction)
	if err = collector.DumpCSV(csvPath+".csv", train, ClassNameToId); err != nil {
		return err
	}
	if err = collector.DumpCSV(csvPath+"-test.csv", test, ClassNameToId); err != nil {
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
