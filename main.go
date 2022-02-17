package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/nordicsense/landsat/correction"
	"github.com/nordicsense/landsat/hist"
	"github.com/nordicsense/landsat/io"
	"github.com/teris-io/cli"
)

// run with e.g. compress=deflate zlevel=6 predictor=3
// best for float32, see https://kokoalberti.com/articles/geotiff-compression-optimization-guide/
func main() {
	correctCmd := cli.NewCommand("correct", "Merge LANSAT bands into a single image applying atmospheric correction").
		WithShortcut("c").
		WithArg(cli.NewArg("args", "GDAL arguments, e.g. compress=deflate zlevel=6 predictor=3").AsOptional()).
		WithOption(cli.NewOption("input", "Input directory (default: current)").WithChar('d')).
		WithOption(cli.NewOption("output", "Output directory (default: same as input)").WithChar('o')).
		WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithAction(correctAction)

	histCmd := cli.NewCommand("hist", "Collect histograms of merged LANDSAT images").
		WithShortcut("h").
		WithOption(cli.NewOption("input", "Input directory (default: current)").WithChar('d')).
		WithOption(cli.NewOption("output", "Output directory (default: same as input)").WithChar('o')).
		WithOption(cli.NewOption("pattern", "Match only files with this pattern (default: .*_T1.tiff$)").WithChar('p')).
		WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithAction(histAction)

	app := cli.New("Tools for processing LANDSAT images").
		WithCommand(correctCmd).
		WithCommand(histCmd)
	
	os.Exit(app.Run(os.Args, os.Stdout))
}

func correctAction(args []string, options map[string]string) int {
	fNames, pathOut, verbose := parseOptions(options, ".*_B1.TIF")
	for _, fName := range fNames {
		pathIn := path.Dir(fName)
		pattern := strings.Replace(path.Base(fName), "_B1.TIF", "", 1)
		if verbose {
			log.Printf("Merging and correcting %s into %s\n", pathIn, pathOut)
		}
		if err := correction.MergeAndApply(pathIn, pattern, pathOut, args...); err != nil {
			log.Fatal(err)
		}
	}
	return 0
}

func histAction(args []string, options map[string]string) int {
	pattern, ok := options["pattern"]
	if !ok {
		pattern = ".*_T1.tiff$"
	}
	fNames, pathOut, verbose := parseOptions(options, pattern)
	for _, fName := range fNames {
		if verbose {
			log.Printf("Collecting histogram for %s into %s", fName, pathOut)
		}
		if err := hist.CollectForMergedImage(fName, pathOut); err != nil {
			log.Fatal(err)
		}
	}
	return 0
}

func parseOptions(options map[string]string, pattern string) ([]string, string, bool) {
	var (
		root, pathOut string
		fNames        []string
		err           error
		ok, verbose   bool
	)
	if root, ok = options["input"]; !ok {
		root, _ = os.Getwd()
	}
	if fNames, err = io.ScanTree(root, pattern); err != nil {
		log.Fatal(err)
	}
	if pathOut, ok = options["output"]; !ok {
		pathOut = root
	}
	if _, ok = options["verbose"]; ok {
		verbose = true
	}
	return fNames, pathOut, verbose
}
