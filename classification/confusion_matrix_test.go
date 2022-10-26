package classification_test

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/nordicsense/landsat/classification"
	"github.com/nordicsense/landsat/data"
)

func TestModelPredict(t *testing.T) {

	root := "/Volumes/Caffeine/Data/Landsat/results/v11"

	data, expected, err := readTestingSet(path.Join(root, "trainingdata", "trainingdata-test.csv"))
	if err != nil {
		t.Error(err)
	}

	model, err := classification.LoadModel(path.Join(root, "tf.model"))
	if err != nil {
		t.Error(err)
	}
	defer model.Close()

	outcomes, err := model.Predict(data)

	var confusionMatrix [classification.NClasses][classification.NClasses]int
	matches := 0
	for i, e := range expected {
		o := outcomes[i]
		confusionMatrix[e][o]++
		if e == o {
			matches++
		}
	}
	accuracy := float64(matches) / float64(len(expected))
	log.Println(confusionMatrix)
	log.Println(accuracy)

}

func readTestingSet(dataFile string) ([]classification.Observation, []int, error) {
	var expected []int
	var xx []classification.Observation

	tfile, err := os.Open(dataFile)
	if err != nil {
		return nil, nil, err
	}
	defer tfile.Close()

	var recs int
	r := csv.NewReader(tfile)
	for {
		rec, err := r.Read()
		if recs == 0 {
			recs++
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var (
			o  int
			vv [data.NVars]float64
		)
		for i, v := range rec {
			if i == 0 {
				continue
			}
			if i == 1 {
				o, err = strconv.Atoi(v)
				if err != nil {
					return nil, nil, err
				}
				continue
			}

			vv[i-2], err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, nil, err
			}
		}
		expected = append(expected, o)
		xx = append(xx, vv)
	}
	return xx, expected, nil
}
