package field

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/nordicsense/landsat/data"
	"github.com/nordicsense/landsat/field/collector"
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
		"dry_tailing_pond":        {"non-veg-forest-dam", 4},
		"residential_area":        {"non-veg-forest-dam", 4},
		"asphalt":                 {"non-veg-forest-dam", 4},
		"quarry":                  {"non-veg-forest-dam", 4},
		"industrial_area":         {"non-veg-forest-dam", 4},
		"human_technogenic_barren_almost_with_no_vegetation": {"non-veg-mix-dam", 5},
		"human_severely_damaged":                             {"non-veg-mix-dam", 5},
		"road":                                               {"non-veg-mix-dam", 5},
		"human_forest_technogenic_barren_with_no_vegetation": {"non-veg-mix-dam", 5},
		"spoil_heap":                                                   {"non-veg-mix-dam", 5},
		"stone_dry_river_in_mountain":                                  {"non-veg-tundra", 6},
		"tundra_stone_tundra":                                          {"non-veg-tundra", 6},
		"agricultural_field_grass_birch_willow":                        {"veg-agriculture", 7},
		"wetland_with_dwarf_shrub_and_open_water":                      {"veg-wetland", 8},
		"natural_undam_grey_willow_with_dwarf_shrub_grass":             {"veg-wetland", 8},
		"wetland_with_dwarf_shrub_grass":                               {"veg-wetland", 8},
		"wetland_turf":                                                 {"veg-wetland", 8},
		"wetland_with_dwarf_shrub_moss_grass":                          {"veg-wetland", 8},
		"wetland_with_grass_moss_dwarf_shrub":                          {"veg-wetland", 8},
		"human_moderately_damaged_spruce_forest":                       {"veg-conif-dam", 9},
		"human_mostly_damaged_birch_spruce":                            {"veg-mix-dam", 10},
		"new_burnt_area":                                               {"veg-mix-newburn", 11},
		"old_burnt_area":                                               {"veg-mix-oldburn", 12},
		"natural_undam_pine_forest_with_dwarf_shrub_and_lichen":        {"veg-conif", 13},
		"natural_undam_spruce_forest_with_dwarf_shrub_and_moss-lichen": {"veg-conif", 13},
		"natural_undam_pine_forest_with_dwarf_shrub_and_moss-lichen":   {"veg-conif", 13},
		"natural_undam_pine_forest_with_dwarf_shrub":                   {"veg-conif", 13},
		"natural_undam_pine_spruce_forest_with_dwarf_shrub":            {"veg-conif", 13},
		"natural_undam_spruce_forest_with_dwarf_shrub":                 {"veg-conif", 13},
		"natural_undam_birch_forest_with_dwarf_shrub_lichen":           {"veg-decid", 14},
		"natural_undam_birch_forest_with_lichen_dwarf_shrub":           {"veg-decid", 14},
		"natural_undam_birch_forest_with_grass":                        {"veg-mix-grass", 15},
		"natural_undam_birch_spruce_forest_with_moss_lichen":           {"veg-mix", 16},
		"tundra_undam_lichen_dwarf_shrub":                              {"veg-tundra", 17},
		"tundra_undam_lichen":                                          {"veg-tundra", 17},
		"tundra_undam_stone_with_lichen":                               {"veg-tundra", 17},
	}

	images = map[string]bool{
		"LE07_L1TP_186012_20000728": true,
		"LE07_L1TP_188012_20000726": false, // false
		"LE07_L1TP_195011_20000727": false, // false
		"LE07_L1TP_195012_20000727": false, // false
		"LT05_L1TP_187012_20050709": true,
		"LT05_L1TP_188012_19860728": true,
		"LT05_L1TP_190011_20090725": true,
		"LT05_L1TP_190012_19930713": false, // false
	}

	r       *rand.Rand
	clazzId map[string]int
)

const NClasses = 18

func init() {
	r = rand.New(rand.NewSource(42))
	clazzId = make(map[string]int)
	for _, m := range mapping {
		clazzId[m.clazz] = m.index
	}
	if NClasses != len(clazzId) {
		panic(fmt.Sprintf("incorrect mapping, got %d classes, want %d", len(clazzId), NClasses))
	}
}

const (
	trainFraction = 0.9
	clazzSize     = 50000
	testSize      = 1000
)

func Collect(tabPath, imgPath, csvPath, imgPattern string) error {
	coord, err := collector.CollectCoordinates(tabPath)
	if err != nil {
		return err
	}
	recs, err := collector.TrainingData(imgPath, imgPattern, coord, convert)
	if err != nil {
		return err
	}
	train, test := collector.Subsample(recs, clazzId, clazzSize, testSize, r, trainFraction)
	if err = collector.DumpCSV(csvPath+".csv", train, clazzId); err != nil {
		return err
	}
	return collector.DumpCSV(csvPath+"-test.csv", test, clazzId)
}

func convert(im string, clazz string, xx []float64) (string, []float64, bool) {
	var ok bool
	newclazz, ok := mapping[clazz]
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
	return newclazz.clazz, data.Transform(xx, landsatId), true
}
