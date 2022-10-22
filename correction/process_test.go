package correction_test

import (
	"github.com/nordicsense/landsat/correction"
	"testing"
)

func TestMergeAndApply(t *testing.T) {
	srcDir := "/Volumes/Caffeine/Data/Landsat/sources/prod/1985.1_LT05_L1TP_188012_19850709_20200918_02_T1"
	pattern := "LT05_L1TP_188012_19850709_20200918_02_T1"
	outDir := "/Volumes/Caffeine/Data/Landsat/corrected"

	err := correction.MergeAndApply(srcDir, pattern, outDir, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
}
