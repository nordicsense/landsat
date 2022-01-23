package field_test

import (
	"encoding/csv"
	"github.com/nordicsense/landsat/field"
	"os"
	"strconv"
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
	res, err := field.TrainingData(imgPathIn, ".*_T1_fix.tiff", coord)
	if err != nil {
		t.Fatal(err)
	}

	fo, err := os.Create("/Users/osklyar/Data/Landsat/analysis/training/training-data-raw-fix.csv")
	if err != nil {
		t.Fatal(err)
	}

	w := csv.NewWriter(fo)

	r := []string{"clazz", "subclazz", "source", "x", "y", "b1", "b2", "b3", "b4", "b5", "b6", "b7"}
	if err = w.Write(r); err != nil {
		t.Fatal(err)
	}
	for _, rr := range res {
		r[0] = rr.Clazz
		r[1] = rr.Subclazz
		r[2] = rr.Source
		r[3] = strconv.Itoa(rr.Coords[0])
		r[4] = strconv.Itoa(rr.Coords[1])
		for i, v := range rr.Bands {
			r[i+5] = strconv.FormatFloat(v, 'f', 4, 64)
		}
		if err = w.Write(r); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	fo.Close()
	t.Log("good")
}
