package trim_test

import (
	"github.com/nordicsense/landsat/trim"
	"testing"
)

func TestProcess(t *testing.T) {
	inputTiff := "/Volumes/Caffeine/Data/Landsat/results/v3-13c-7v/LC08_L1TP_187012_20210705_20210713_02_T1.tiff"
	trim.Process(inputTiff, "", true, true, trim.TL, trim.TR, trim.BR, trim.BL)
}
