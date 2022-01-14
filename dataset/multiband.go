package dataset

import (
	"fmt"

	"github.com/nordicsense/gdal"
)

func OpenMultiBand(fileName string) (MultiBandReader, error) {
	ds, err := gdal.Open(fileName, gdal.ReadOnly)
	if err != nil {
		return nil, err
	}
	nb := ds.RasterCount()
	if nb < 1 {
		ds.Close()
		return nil, fmt.Errorf("no raster bands found")
	}
	var bands []*uniBand
	var ip *ImageParams
	for i := 1; i <= nb; i++ {
		band, err := openSingleBand(ds, i)
		if err != nil {
			ds.Close()
			return nil, err
		}
		bands = append(bands, band)
		if i == 1 {
			ip = band.ImageParams()
		}
	}
	return &multiBand{Dataset: ds, ip: ip, bands: bands}, nil
}

func NewMultiBand(fileName string, driver Driver, n int, ip *ImageParams, options ...string) (MultiBandWriter, error) {
	gdalDriver, err := gdal.GetDriverByName(string(driver))
	if err != nil {
		return nil, err
	}
	ds := gdalDriver.Create(fileName, ip.XSize(), ip.YSize(), n, ip.DataType(), options)
	if err = ds.SetGeoTransform(ip.Transform()); err != nil {
		return nil, err
	}
	if err = ds.SetProjection(ip.Projection()); err != nil {
		return nil, err
	}
	if nan, hasnan := ip.NaN(); hasnan {
		if err := ds.RasterBand(1).SetNoDataValue(nan); err != nil {
			return nil, err
		}
	}
	var bands []*uniBand
	for i := 1; i <= ds.RasterCount(); i++ {
		bands = append(bands, &uniBand{Dataset: ds, band: i, ip: ip, rp: &RasterParams{}})
	}
	return &multiBand{Dataset: ds, ip: ip, bands: bands}, nil
}

type multiBand struct {
	gdal.Dataset
	ip *ImageParams
	bands []*uniBand
}

func (mb *multiBand) ImageParams() *ImageParams {
	return mb.ip
}

func (mb *multiBand) Bands() int {
	return mb.RasterCount()
}

func (mb *multiBand) Reader(band int) UniBandReader {
	if band > mb.RasterCount() {
		return nil
	}
	return mb.bands[band-1]
}

func (mb *multiBand) Writer(band int) UniBandWriter {
	if band > mb.RasterCount() {
		return nil
	}
	return mb.bands[band-1]
}

func (mb *multiBand) Close() {
	mb.bands = nil
	mb.ip = nil
	mb.Dataset.Close()
}
