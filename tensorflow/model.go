package tensorflow

import (
	"fmt"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

const (
	NVariables = 10
	NClasses   = 13

	defaultModelTag = "serve"
	modelInputOp    = "serving_default_outer_input"
	modelOutputOp   = "StatefulPartitionedCall"
)

type Observation [NVariables]float64

func LoadModel(name string) (*Model, error) {
	model, err := tf.LoadSavedModel(name, []string{defaultModelTag}, nil)
	if err != nil {
		return nil, err
	}
	return &Model{m: model}, nil
}

type Model struct {
	m *tf.SavedModel
}

func (m *Model) Predict(obs []Observation) ([]int, error) {
	if len(obs) < 1 {
		return nil, nil
	}
	tensor, err := tf.NewTensor(obs)
	if err != nil {
		return nil, err
	}

	feeds := map[tf.Output]*tf.Tensor{m.m.Graph.Operation(modelInputOp).Output(0): tensor}
	fetches := []tf.Output{m.m.Graph.Operation(modelOutputOp).Output(0)}

	output, err := m.m.Session.Run(feeds, fetches, nil)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, fmt.Errorf("empty output")
	}
	outdata, ok := (output[0].Value()).([][]float32)
	if !ok {
		return nil, fmt.Errorf("unexpected type of output: %T", output[0].Value())
	}
	if len(outdata) != len(obs) {
		return nil, fmt.Errorf("incorrect size of output: expected %d, found %d", len(obs), len(outdata))
	}

	res := make([]int, len(outdata))
	for i, x := range outdata {
		res[i] = indexOfMaxValue(x)
	}
	return res, nil
}

func indexOfMaxValue(x []float32) int {
	var max float32
	ind := -1
	for j := 0; j < len(x); j++ {
		if x[j] > max {
			max = x[j]
			ind = j
		}
	}
	return ind
}

func (m *Model) Close() error {
	err := m.m.Session.Close()
	m.m = nil
	return err
}
