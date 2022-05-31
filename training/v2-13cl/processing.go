package field

import (
	"math"
	"math/rand"

	"github.com/nordicsense/landsat/data"
	"github.com/nordicsense/landsat/training/collector"
)

var (
	mapping = map[string]string{
		"human_technogenic_barren_almost_with_no_vegetation": "nonvegetated",
		"stone_dry_river_in_mountain":                        "nonvegetated",
		"dry_tailing_pond":                                   "nonvegetated",
		"asphalt":                                            "nonvegetated",
		"quarry":                                             "nonvegetated",
		"tundra_stone_tundra":                                "nonvegetated",
		"human_forest_technogenic_barren_with_no_vegetation": "nonvegetated",
		"tundra_undam_stone_with_lichen":                     "nonvegetated",

		"residential_area":                                   "impact-nonvegetated",
		"wet_tailing_pond":                                   "impact-nonvegetated",
		"road":                                               "impact-nonvegetated",
		"industrial_area":                                    "impact-nonvegetated",
		"spoil_heap":                                         "impact-nonvegetated",
		"human_moderately_damaged_spruce_forest":             "impact-damaged",
		"human_mostly_damaged_birch_spruce":                  "impact-damaged",
		"human_severely_damaged":                             "impact-damaged",
		"industrial_water":                                   "impact-water",
		"water_with_sediments":                               "impact-water",
		"very_wet_tailing_pond":                              "impact-water",
		"water_with_no_sediments":                            "water",
		"cloud":                                              "cloud",
		"tundra_undam_lichen":                                "tundra",
		"tundra_undam_lichen_dwarf_shrub":                    "tundra",
		"old_burnt_area":                                     "burnt-old",
		"new_burnt_area":                                     "burnt-new",
		"agricultural_field_grass_birch_willow":              "agricultural",
		"wetland_with_dwarf_shrub_and_open_water":            "wetland",
		"wetland_with_dwarf_shrub_grass":                     "wetland",
		"wetland_turf":                                       "wetland",
		"wetland_with_dwarf_shrub_moss_grass":                "wetland",
		"wetland_with_grass_moss_dwarf_shrub":                "wetland",
		"natural_undam_birch_spruce_forest_with_moss_lichen": "decidious",
		"natural_undam_birch_forest_with_dwarf_shrub_lichen": "decidious",
		"natural_undam_birch_forest_with_lichen_dwarf_shrub": "decidious",
		"natural_undam_grey_willow_with_dwarf_shrub_grass":   "decidious",
		"natural_undam_birch_forest_with_grass":              "decidious",

		"natural_undam_spruce_forest_with_dwarf_shrub_and_moss-lichen": "coniferous",
		"natural_undam_pine_forest_with_dwarf_shrub_and_lichen":        "coniferous",
		"natural_undam_pine_forest_with_dwarf_shrub_and_moss-lichen":   "coniferous",
		"natural_undam_pine_forest_with_dwarf_shrub":                   "coniferous",
		"natural_undam_pine_spruce_forest_with_dwarf_shrub":            "coniferous",
		"natural_undam_spruce_forest_with_dwarf_shrub":                 "coniferous",
	}

	clazzId = map[string]int{
		"cloud":               0,
		"water":               1,
		"impact-water":        2,
		"agricultural":        3,
		"burnt-new":           4,
		"burnt-old":           5,
		"impact-damaged":      6,
		"impact-nonvegetated": 7,
		"nonvegetated":        8,
		"tundra":              9,
		"wetland":             10,
		"coniferous":          11,
		"decidious":           12,
	}

	images = map[string]bool{
		"LE07_L1TP_186012_20000728": true,
		"LE07_L1TP_188012_20000726": false, // false 0.9, 0.70
		"LE07_L1TP_195011_20000727": false, // false 0.78, 0.74
		"LE07_L1TP_195012_20000727": false, // false, 0.83, 0.72
		"LT05_L1TP_187012_20050709": true,
		"LT05_L1TP_188012_19860728": true,
		"LT05_L1TP_190011_20090725": true, // false, 0.47, 0.84
		"LT05_L1TP_190012_19930713": false,
	}

	r = rand.New(rand.NewSource(42))
)

const (
	NClasses = 13

	trainFraction = 0.8
	clazzSize     = 40000
	testSize      = 500
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
	return newclazz, data.Transform(xx, landsatId), true
}
