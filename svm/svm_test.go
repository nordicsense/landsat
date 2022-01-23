package svm_test

import (
	"log"
	"testing"

	"github.com/nordicsense/landsat/field"
	"github.com/nordicsense/landsat/svm"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func TestProcess(t *testing.T) {
	fieldDataPathIn := "/Users/osklyar/Data/Landsat/TrainingSet"
	imgPathIn := "/Users/osklyar/Data/Landsat/analysis/training"
	coord, err := field.Coordinates(fieldDataPathIn)
	if err != nil {
		t.Fatal(err)
	}
	//data, err := field.TrainingData(imgPathIn, ".*_T1_fix.tiff", coord)
	data, err := field.TrainingData(imgPathIn, ".*_T1.tiff", coord)
	if err != nil {
		t.Fatal(err)
	}

	svm.Process(data)
}
