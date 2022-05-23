package field_test

import (
	"os"
	"path"
	"testing"

	"github.com/nordicsense/landsat/field"
)

func TestCollect(t *testing.T) {
	hd, _ := os.UserHomeDir()
	err := field.Collect(
		path.Join(hd, "Data/Landsat/TrainingSet"),
		path.Join(hd, "Data/Landsat/analysis/training"),
		path.Join(hd, "Data/Landsat/analysis/model/trainingdata"),
		".*.tiff")
	if err != nil {
		t.Fatal(err)
	}
}
