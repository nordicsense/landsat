package field_test

import (
	"github.com/nordicsense/landsat/field"
	"testing"
)

func TestCoordinates(t *testing.T) {
	pathIn := "/Users/osklyar/Data/Landsat/TrainingSet"
	res, err := field.Coordinates(pathIn)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestTrainingData(t *testing.T) {
	fieldDataPathIn := "/Users/osklyar/Data/Landsat/TrainingSet"
	imgPathIn := "/Users/osklyar/Data/Landsat/analysis/training"
	coord, err := field.Coordinates(fieldDataPathIn)
	if err != nil {
		t.Fatal(err)
	}
	res, err := field.TrainingData(imgPathIn, ".*_T1.tiff", coord)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
