package tensorflow_test

import (
	"github.com/nordicsense/landsat/tensorflow"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestPredict(t *testing.T) {
	_, testFileName, _, _ := runtime.Caller(0)
	assetsDir := path.Join(path.Dir(testFileName), "test_assets")

	hd, _ := os.UserHomeDir()

	modelDir := path.Join(assetsDir, "tf.model")
	inputTiff := path.Join(hd, "Data/Landsat/analysis/prod/LE07_L1TP_186012_20000728_20200918_02_T1.tiff")
	outputTiff := path.Join(hd, "Data/Landsat/analysis/LE07_L1TP_186012_20000728_20200918_02_T1-classification.tiff")

	err := tensorflow.Predict(modelDir, inputTiff, outputTiff, 0, 9000, 0, 9000, false, false)
	if err != nil {
		t.Fatal(err)
	}
}
