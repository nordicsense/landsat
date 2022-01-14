package dataset

import (
	"fmt"
	"github.com/nordicsense/gdal"
	"math"
)

func OpenUniBand(fileName string) (UniBandReader, error) {
	ds, err := gdal.Open(fileName, gdal.ReadOnly)
	if err != nil {
		return nil, err
	}
	if ds.RasterCount() < 1 {
		ds.Close()
		return nil, fmt.Errorf("no raster bands found")
	}
	return openSingleBand(ds, 1)
}

func openSingleBand(ds gdal.Dataset, band int) (*uniBand, error) {
	rb := ds.RasterBand(band)
	ipb := ImageParamsBuilder(ds.RasterXSize(), ds.RasterYSize()).
		DataType(rb.RasterDataType()).
		Transform(ds.GeoTransform()).
		Projection(ds.Projection())
	if nan, ok := rb.NoDataValue(); ok {
		ipb = ipb.NaN(nan)
	}
	rpb := RasterParamsBuilder()
	if scale, ok := rb.GetScale(); ok {
		rpb = rpb.Scale(scale)
	}
	if offset, ok := rb.GetOffset(); ok {
		rpb = rpb.Offset(offset)
	}
	for _, k := range rb.Metadata(domain) {
		rpb = rpb.Metadata(k, rb.MetadataItem(k, domain))
	}
	return &uniBand{Dataset: ds, band: band, ip: ipb.Build(), rp: rpb.Build()}, nil
}

func NewUniBand(fileName string, driver Driver, ip *ImageParams, rp *RasterParams, options ...string) (UniBandWriter, error) {
	gdalDriver, err := gdal.GetDriverByName(string(driver))
	if err != nil {
		return nil, err
	}
	ds := gdalDriver.Create(fileName, ip.XSize(), ip.YSize(), /* bands= */1, ip.DataType(), options)
	if err = ds.SetGeoTransform(ip.Transform()); err != nil {
		return nil, err
	}
	if err = ds.SetProjection(ip.Projection()); err != nil {
		return nil, err
	}
	ub := &uniBand{Dataset: ds, band: 1, ip: ip}
	return ub, ub.SetRasterParams(rp)
}

type uniBand struct {
	gdal.Dataset
	band int
	ip   *ImageParams
	rp   *RasterParams
}

func (ub *uniBand) ImageParams() *ImageParams {
	return ub.ip
}

func (ub *uniBand) RasterParams() *RasterParams {
	return ub.rp
}

func (ub *uniBand) SetRasterParams(rp *RasterParams) error {
	ub.rp = rp
	var err error
	rb := ub.RasterBand(ub.band)
	if err = rb.SetOffset(rp.Offset()); err != nil {
		return err
	}
	if err = rb.SetScale(rp.Scale()); err != nil {
		return err
	}
	for k, v := range rp.Metadata() {
		if err := rb.SetMetadataItem(k, v, domain); err != nil {
			return err
		}
	}
	return err
}

func (ub *uniBand) Read(x, y int) (float64, error) {
	nx := ub.ImageParams().XSize()
	ny := ub.ImageParams().YSize()
	if x < 0 || x >= nx || y < 0 || y >= ny {
		return math.NaN(), fmt.Errorf("{x:%d, y:%d} is outside of image area {x:[0,%d), y:[,%d)}", x, y, nx, ny)
	}
	if res, err := ub.ReadBlock(x, y, Box{0, 0, 1, 1}); err == nil {
		return res[0], nil
	} else {
		return math.NaN(), err
	}
}

func (ub *uniBand) ReadAtLatLon(ll LatLon) (float64, error) {
	x, y := ub.ImageParams().Transform().LatLon2Pixels(ll)
	return ub.Read(x, y)
}

func (ub *uniBand) ReadBlock(x, y int, box Box) ([]float64, error) {
	rb := ub.Dataset.RasterBand(ub.band)
	buffer := make([]float64, box[2]*box[3])
	err := rb.IO(gdal.Read, x+box[0], y+box[1], box[2], box[3], buffer, box[2], box[3], 0, 0)
	if err != nil {
		return nil, err
	}
	ip := ub.ImageParams()
	rp := ub.RasterParams()
	nan, hasnan := ip.NaN()
	for i, val := range buffer {
		if hasnan && buffer[i] == nan {
			buffer[i] = math.NaN()
		} else {
			buffer[i] = val*rp.Scale() + rp.Offset()
		}
	}
	return buffer, nil
}

func (ub *uniBand) Write(x, y int, v float64) error {
	return ub.WriteBlock(x, y, Box{0, 0, 1, 1}, []float64{v})
}

func (ub *uniBand) WriteAtLatLon(ll LatLon, v float64) error {
	x, y := ub.ImageParams().Transform().LatLon2Pixels(ll)
	return ub.Write(x, y, v)
}

func (ub *uniBand) WriteBlock(x, y int, box Box, buffer []float64) error {
	rb := ub.Dataset.RasterBand(ub.band) // Assume 1 band or panic
	// GDAL can handle any format, but it is more efficient to use specific type as we need to make a copy anyway
	p := ub.RasterParams()
	switch ub.ImageParams().DataType() {
	case gdal.Int32:
		data := make([]int32, len(buffer))
		for i, v := range buffer {
			data[i] = int32((v - p.Offset()) / p.Scale())
		}
		return rb.IO(gdal.Write, x+box[0], y+box[1], box[2], box[3], data, box[2], box[3], 0, 0)
	case gdal.Float32:
		data := make([]float32, len(buffer))
		for i, v := range buffer {
			data[i] = float32((v - p.Offset()) / p.Scale())
		}
		return rb.IO(gdal.Write, x+box[0], y+box[1], box[2], box[3], data, box[2], box[3], 0, 0)
	default: // treat as float64
		data := make([]float64, len(buffer))
		for i, v := range buffer {
			data[i] = (v - p.Offset()) / p.Scale()
		}
		return rb.IO(gdal.Write, x+box[0], y+box[1], box[2], box[3], data, box[2], box[3], 0, 0)
	}
}

func (ub *uniBand) BreakGlass() gdal.Dataset {
	return ub.Dataset
}

func (ub *uniBand) Close() {
	ub.Dataset.Close()
	ub.ip = nil
}
