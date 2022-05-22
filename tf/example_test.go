package tf_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"testing"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

func TestTf(t *testing.T) {
	// Construct a graph with an operation that produces a string constant.
	s := op.NewScope()
	c := op.Const(s, "Hello from TensorFlow version "+tf.Version())
	graph, err := s.Finalize()
	if err != nil {
		panic(err)
	}

	// Execute the graph in a session.
	sess, err := tf.NewSession(graph, nil)
	if err != nil {
		panic(err)
	}
	output, err := sess.Run(nil, []tf.Output{c}, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(output[0].Value())
}

func TestPredict(t *testing.T) {
	model, err := tf.LoadSavedModel("/Users/osklyar/Data/Landsat/analysis/model/tf.model",
		[]string{"serve"}, nil)
	if err != nil {
		t.Error(err)
	}
	defer model.Session.Close()

	tfile, err := os.Open("/Users/osklyar/Data/Landsat/analysis/model/trainingdata-test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer tfile.Close()

	var expected []int
	var data [][10]float64
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
			vv [10]float64
		)
		for i, v := range rec {
			if i == 0 {
				continue
			}
			if i == 1 {
				o, err = strconv.Atoi(v)
				if err != nil {
					t.Fatal(err)
				}
				continue
			}

			vv[i-2], err = strconv.ParseFloat(v, 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		expected = append(expected, o)
		data = append(data, vv)
	}

	tensor, err := tf.NewTensor(data)
	if err != nil {
		t.Fatal(err)
	}

	feeds := map[tf.Output]*tf.Tensor{
		model.Graph.Operation("serving_default_outer_input").Output(0): tensor, // Replace this with your input layer name
	}
	outputs := []tf.Output{
		model.Graph.Operation("StatefulPartitionedCall").Output(0), // Replace this with your output layer name
	}
	res, err := model.Session.Run(feeds, outputs, nil)
	if err != nil {
		t.Fatal(err)
	}
	xx := (res[0].Value()).([][]float32)

	var outcomes []int
	for _, x := range xx {
		var indmax float32
		ind := -1
		for j := 0; j < len(x); j++ {
			if x[j] > indmax {
				indmax = x[j]
				ind = j
			}
		}
		outcomes = append(outcomes, ind)
	}

	var contmatrix [13][13]int
	matches := 0
	for i, e := range expected {
		o := outcomes[i]
		contmatrix[e][o]++
		if e == o {
			matches++
		}
	}
	log.Println(contmatrix)
	log.Println(float64(matches) / float64(len(expected)))
}
