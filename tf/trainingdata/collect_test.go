package trainingdata_test

import (
	"github.com/nordicsense/landsat/tf/trainingdata"
	"testing"
)

func TestCollect(t *testing.T) {
	err := trainingdata.Collect("/Users/osklyar/Data/Landsat/TrainingSet",
		"/Users/osklyar/Data/Landsat/analysis/training", "/Users/osklyar/Data/Landsat/analysis/model/trainingdata")
	if err != nil {
		t.Fatal(err)
	}
}
