package change_test

import (
	"github.com/nordicsense/landsat/change"
	"testing"
)

func TestCollect(t *testing.T) {
	fromDiffs := []string{
		"/Volumes/Caffeine/Data/Landsat/results/v6-11c-7v/trimmed/LC08_L1TP_187012_20140718_20200911_02_T1.tiff",
		"/Volumes/Caffeine/Data/Landsat/results/v6-11c-7v/trimmed/LC08_L1TP_187013_20140718_20200911_02_T1.tiff",
	}
	toDiffs := []string{
		"/Volumes/Caffeine/Data/Landsat/results/v6-11c-7v/trimmed/LC08_L1TP_187012_20170710_20200903_02_T1.tiff",
		"/Volumes/Caffeine/Data/Landsat/results/v6-11c-7v/trimmed/LC08_L1TP_187013_20170710_20200903_02_T1.tiff",
	}

	err := change.Collect(fromDiffs, toDiffs, change.TL, change.BR, "/Volumes/Caffeine/Data/Landsat/results/test")
	if err != nil {
		t.Fatal(err)
	}
}
