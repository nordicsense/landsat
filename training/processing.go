package training

import (
	"fmt"
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
		//"snow":                    {"snow", 0},
		"cloud":                   {"cloud", 0},
		"water_with_no_sediments": {"water", 1},
		"wet_tailing_pond":        {"water-dam", 2},
		"water_with_sediments":    {"water-dam", 2},
		"very_wet_tailing_pond":   {"water-dam", 2},
		"industrial_water":        {"water-dam", 2},
		"dry_tailing_pond":        {"non-veg-dam", 3},
		"residential_area":        {"non-veg-dam", 3},
		"asphalt":                 {"non-veg-dam", 3},
		"quarry":                  {"non-veg-dam", 3},
		"industrial_area":         {"non-veg-dam", 3},
		"human_technogenic_barren_almost_with_no_vegetation": {"non-veg-dam", 3},
		"human_severely_damaged":                             {"non-veg-dam", 3},
		"road":                                               {"non-veg-dam", 3},
		"human_forest_technogenic_barren_with_no_vegetation": {"non-veg-dam", 3},
		"spoil_heap":                                                   {"non-veg-dam", 3},
		"stone_dry_river_in_mountain":                                  {"non-veg-dam", 3},
		"tundra_stone_tundra":                                          {"non-veg-dam", 3},
		"agricultural_field_grass_birch_willow":                        {"veg-dam", 4}, //separate
		"wetland_with_dwarf_shrub_and_open_water":                      {"veg-dam", 4},
		"natural_undam_grey_willow_with_dwarf_shrub_grass":             {"veg-dam", 4},
		"wetland_with_dwarf_shrub_grass":                               {"veg-dam", 4},
		"wetland_turf":                                                 {"veg-dam", 4},
		"wetland_with_dwarf_shrub_moss_grass":                          {"veg-dam", 4},
		"wetland_with_grass_moss_dwarf_shrub":                          {"veg-dam", 4},
		"human_moderately_damaged_spruce_forest":                       {"veg-dam", 4},
		"human_mostly_damaged_birch_spruce":                            {"veg-dam", 4},
		"new_burnt_area":                                               {"veg-newburn", 5},
		"old_burnt_area":                                               {"veg-oldburn", 6},
		"natural_undam_pine_forest_with_dwarf_shrub_and_lichen":        {"veg-pine", 7},
		"natural_undam_spruce_forest_with_dwarf_shrub_and_moss-lichen": {"veg-spruce", 8},
		"natural_undam_pine_forest_with_dwarf_shrub_and_moss-lichen":   {"veg-pine", 7},
		"natural_undam_pine_forest_with_dwarf_shrub":                   {"veg-pine", 7},
		"natural_undam_pine_spruce_forest_with_dwarf_shrub":            {"veg-pine", 7},
		"natural_undam_spruce_forest_with_dwarf_shrub":                 {"veg-spruce", 8},
		"natural_undam_birch_forest_with_dwarf_shrub_lichen":           {"veg-decid", 9},
		"natural_undam_birch_forest_with_lichen_dwarf_shrub":           {"veg-decid", 9},
		"natural_undam_birch_forest_with_grass":                        {"veg-dam", 4}, // separate
		//"natural_undam_birch_spruce_forest_with_moss_lichen":           {"veg-nonforest", -1},
		"tundra_undam_lichen_dwarf_shrub": {"veg-tundra", 10},
		"tundra_undam_lichen":             {"veg-tundra", 10},
		"tundra_undam_stone_with_lichen":  {"veg-tundra", 10},
	}

	images = map[string]bool{
		"LT05_L1TP_190011_20090725": true,
		"LT05_L1TP_187012_20050709": true,
		"LE07_L1TP_186012_20000728": true,
		"LT05_L1TP_188012_19860728": true,
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
	trainFraction = 0.9
	clazzSize     = 40000
	testSize      = 2000
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
	return collector.DumpCSV(csvPath+"-test.csv", test, ClassNameToId)
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
