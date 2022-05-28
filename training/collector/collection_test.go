package collector_test

import (
	"encoding/csv"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/nordicsense/landsat/training/collector"
)

func TestCoordinates(t *testing.T) {
	hd, _ := os.UserHomeDir()

	pathIn := path.Join(hd, "Data/Landsat/TrainingSet")
	res, err := collector.CollectCoordinates(pathIn)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestTrainingData(t *testing.T) {
	hd, _ := os.UserHomeDir()
	fieldDataPathIn := path.Join(hd, "Data/Landsat/TrainingSet")
	imgPathIn := path.Join(hd, "Data/Landsat/analysis/training")
	coord, err := collector.CollectCoordinates(fieldDataPathIn)
	if err != nil {
		t.Fatal(err)
	}
	res, err := collector.TrainingData(imgPathIn, ".*_T1.tiff", coord, collector.PathThrough)
	if err != nil {
		t.Fatal(err)
	}

	fo, err := os.Create(path.Join(hd, "Data/Landsat/analysis/training/training-data-raw-fix.csv"))
	if err != nil {
		t.Fatal(err)
	}

	w := csv.NewWriter(fo)

	r := []string{"clazz", "source", "x", "y", "b1", "b2", "b3", "b4", "b5", "b6", "b7"}
	if err = w.Write(r); err != nil {
		t.Fatal(err)
	}
	for _, rr := range res {
		r[0] = rr.Clazz
		r[1] = rr.Image
		r[2] = strconv.Itoa(rr.Coords[0])
		r[3] = strconv.Itoa(rr.Coords[1])
		for i, v := range rr.Data {
			r[i+4] = strconv.FormatFloat(v, 'f', 4, 64)
		}
		if err = w.Write(r); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	fo.Close()
	t.Log("good")
}