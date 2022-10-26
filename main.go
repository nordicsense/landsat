package main

import (
	"github.com/nordicsense/landsat/classification"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nordicsense/landsat/change"
	"github.com/nordicsense/landsat/conversion"
	"github.com/nordicsense/landsat/filter"
	"github.com/nordicsense/landsat/io"
	"github.com/nordicsense/landsat/trim"
	"github.com/teris-io/cli"
)

// run with e.g. compress=deflate zlevel=6 predictor=3
// best for float32, see https://kokoalberti.com/articles/geotiff-compression-optimization-guide/
func main() {
	convertCmd := cli.NewCommand("convert", "Merge LANDSAT bands into a single image").
		WithShortcut("c").
		WithArg(cli.NewArg("args", "GDAL arguments, e.g. compress=deflate zlevel=6 predictor=3").AsOptional()).
		WithOption(cli.NewOption("input", "Input directory (default: current)").WithChar('d')).
		WithOption(cli.NewOption("output", "Output directory (default: same as input)").WithChar('o')).
		WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithOption(cli.NewOption("l1", "L1 (default: L2, off)").WithChar('l').WithType(cli.TypeBool)).
		WithOption(cli.NewOption("skip", "Skip existing").WithChar('s').WithType(cli.TypeBool)).
		WithAction(convertAction)

	trainingCmd := cli.NewCommand("training", "Collect training data from field data").
		WithShortcut("t").
		WithArg(cli.NewArg("coorddir", "Directory with coordinate files")).
		WithOption(cli.NewOption("input", "Input directory for images (default: current)").WithChar('d')).
		WithOption(cli.NewOption("output", "Output directory for training data (default: current)").WithChar('o')).
		// WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithAction(fieldDataAction)

	predictCmd := cli.NewCommand("predict", "Predict land cover classes with Tensorflow classification").
		WithShortcut("p").
		WithArg(cli.NewArg("data", "Multi-band Landsat GeoTiff with 7 bands of input data")).
		WithOption(cli.NewOption("model", "Tensorflow model directory (default: ./tf.model)").WithChar('m')).
		WithOption(cli.NewOption("output", "Output directory (default: same as input)").WithChar('o')).
		WithOption(cli.NewOption("id", "Landsat series Id (5, 7 (default), or 8)").WithType(cli.TypeInt)).
		WithOption(cli.NewOption("skip", "Skip existing").WithChar('s').WithType(cli.TypeBool)).
		WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithAction(predictAction)

	filterCmd := cli.NewCommand("filter", "Filter output with a smoothing filter").
		WithShortcut("f").
		WithArg(cli.NewArg("algo", "Filtering algorithm: 3x3, 5x5")).
		WithArg(cli.NewArg("data", "Classification uni-band")).
		WithOption(cli.NewOption("output", "Output directory (default: same as input)").WithChar('o')).
		WithOption(cli.NewOption("skip", "Skip existing").WithChar('s').WithType(cli.TypeBool)).
		WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithAction(filterAction)

	trimCmd := cli.NewCommand("trim", "Filter output with a smoothing filter").
		WithArg(cli.NewArg("data", "Image to trim")).
		WithOption(cli.NewOption("output", "Output directory (default: same as input)").WithChar('o')).
		WithOption(cli.NewOption("skip", "Skip existing").WithChar('s').WithType(cli.TypeBool)).
		WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithAction(trimAction)

	changeCmd := cli.NewCommand("change", "Change detection").
		WithArg(cli.NewArg("from", "2 from images")).
		WithArg(cli.NewArg("to", "2 to images")).
		WithOption(cli.NewOption("output", "Output directory (default: same as input)").WithChar('o')).
		WithAction(changeAction)

	app := cli.New("Normalize and classify Landsat images for the Northern hemisphere").
		WithCommand(convertCmd).
		WithCommand(trainingCmd).
		WithCommand(predictCmd).
		WithCommand(filterCmd).
		WithCommand(trimCmd).
		WithCommand(changeCmd)

	os.Exit(app.Run(os.Args, os.Stdout))
}

func convertAction(args []string, options map[string]string) int {
	var (
		ok, skip, l1 bool
		err          error
		root         string
		fNames       []string
	)
	if root, ok = options["input"]; !ok {
		root, _ = os.Getwd()
	}
	if fNames, err = io.ScanTree(root, ".*_B1.TIF"); err != nil {
		log.Fatal(err)
	}
	if _, ok = options["l1"]; ok {
		l1 = true
	}
	if _, ok = options["skip"]; ok {
		skip = true
	}
	pathOut, verbose := parseOptions(root, options)
	for _, fName := range fNames {
		pathIn := path.Dir(fName)
		pattern := strings.Replace(path.Base(fName), "_B1.TIF", "", 1)
		if verbose {
			log.Printf("Merging and correcting %s into %s\n", pathIn, pathOut)
		}
		if err := conversion.MergeAndApply(pathIn, pattern, pathOut, l1, skip, verbose, args...); err != nil {
			log.Fatal(err)
		}
	}
	return 0
}

func fieldDataAction(args []string, options map[string]string) int {
	var (
		ok       bool
		imageDir string
	)
	coordDir := args[0]
	current, _ := os.Getwd()
	if imageDir, ok = options["input"]; !ok {
		imageDir = current
	}
	pathOut, _ := parseOptions(current, options)
	if err := classification.CollectTrainingData(coordDir, imageDir, pathOut, ".*.tiff"); err != nil {
		log.Fatal(err)
	}
	return 0
}

func predictAction(args []string, options map[string]string) int {
	var (
		ok       bool
		skip     bool
		modelDir string
	)
	fileIn := args[0]
	if modelDir, ok = options["model"]; !ok {
		current, _ := os.Getwd()
		modelDir = path.Join(current, "tf.model")
	}
	pathOut, verbose := parseOptions(path.Dir(fileIn), options)
	if pathOut == path.Dir(fileIn) {
		pathOut = path.Join(pathOut, "classification")
	}
	_ = os.MkdirAll(pathOut, 0750)
	fileOut := path.Join(pathOut, path.Base(fileIn))
	id := 7
	if idStr, ok := options["id"]; ok {
		id, _ = strconv.Atoi(idStr)
	}
	if _, ok = options["skip"]; ok {
		skip = true
	}
	if err := classification.Predict(modelDir, fileIn, fileOut, 0, 9000, 0, 9000, id, skip, verbose); err != nil {
		log.Fatal(err)
	}
	return 0
}

func filterAction(args []string, options map[string]string) int {
	var (
		ok   bool
		skip bool
	)
	algo := args[0]
	fileIn := args[1]
	pathOut, verbose := parseOptions(path.Dir(fileIn), options)
	pathOut = path.Join(pathOut, algo)
	_ = os.MkdirAll(pathOut, 0750)

	fileOut := path.Join(pathOut, path.Base(fileIn))
	if _, ok = options["skip"]; ok {
		skip = true
	}
	switch algo {
	case "3x3":
		if err := filter.Filter3x3(fileIn, fileOut, skip, verbose); err != nil {
			log.Fatal(err)
		}
	case "5x5":
		if err := filter.Filter5x5(fileIn, fileOut, skip, verbose); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unknown algorithm")
	}

	return 0
}

func trimAction(args []string, options map[string]string) int {
	var (
		ok   bool
		skip bool
	)
	fileIn := args[0]
	pathOut, verbose := parseOptions(path.Dir(fileIn), options)
	pathOut = path.Join(pathOut, "trimmed")
	_ = os.MkdirAll(pathOut, 0750)

	fileOut := path.Join(pathOut, path.Base(fileIn))
	if _, ok = options["skip"]; ok {
		skip = true
	}
	if err := trim.Process(fileIn, fileOut, skip, verbose, trim.TL, trim.TR, trim.BR, trim.BL); err != nil {
		log.Fatal(err)
	}
	return 0
}

func changeAction(args []string, options map[string]string) int {
	fromTiffs := strings.Split(args[0], ",")
	toTiffs := strings.Split(args[1], ",")
	output, ok := options["output"]
	if !ok {
		output, _ = os.Getwd()
	}
	if err := change.Collect(fromTiffs, toTiffs, change.TL, change.BR, output); err != nil {
		log.Fatal(err)
	}
	return 0

}

func parseOptions(root string, options map[string]string) (string, bool) {
	var (
		pathOut     string
		ok, verbose bool
	)
	if pathOut, ok = options["output"]; !ok {
		pathOut = root
	}
	if _, ok = options["verbose"]; ok {
		verbose = true
	}
	return pathOut, verbose
}
