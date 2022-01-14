package dataset

import "github.com/nordicsense/gdal"

type Driver string

// Box defines an area of raster: x,y offset and x,y size.
type Box [4]int

const (
	GTiff Driver = "GTiff"
	domain = ""
)

type UniBandReader interface {
	ImageParams() *ImageParams
	RasterParams() *RasterParams
	Read(x, y int) (float64, error)
	ReadAtLatLon(ll LatLon) (float64, error)
	ReadBlock(x, y int, box Box) ([]float64, error)
	Close()
	BreakGlass() gdal.Dataset
}

type UniBandWriter interface {
	ImageParams() *ImageParams
	RasterParams() *RasterParams
	SetRasterParams(rp *RasterParams) error
	Write(x, y int, v float64) error
	WriteAtLatLon(ll LatLon, v float64) error
	WriteBlock(x, y int, box Box, buffer []float64) error
	Close()
	BreakGlass() gdal.Dataset
}

type MultiBandReader interface {
	ImageParams() *ImageParams
	Bands() int
	Reader(band int) UniBandReader
	Close()
}

type MultiBandWriter interface {
	ImageParams() *ImageParams
	Bands() int
	Writer(band int) UniBandWriter
	Close()
}
