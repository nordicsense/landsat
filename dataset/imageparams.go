package dataset

import (
	"github.com/nordicsense/gdal"
	"math"
)

type ImageParams struct {
	xSize      int
	ySize      int
	transform  AffineTransform
	projection string
	datatype   gdal.DataType
	nan        float64
	nanPresent bool
}

func (p *ImageParams) copy() *ImageParams {
	return &ImageParams{
		xSize:      p.xSize,
		ySize:      p.ySize,
		transform:  p.transform,
		projection: p.projection,
		datatype:   p.datatype,
		nan:        p.nan,
		nanPresent: p.nanPresent,
	}
}

func (p *ImageParams) ToBuilder() *imageParamsBuilder {
	return &imageParamsBuilder{ImageParams: p.copy()}
}

func (p *ImageParams) XSize() int {
	return p.xSize
}

func (p *ImageParams) YSize() int {
	return p.ySize
}

func (p *ImageParams) Transform() AffineTransform {
	return p.transform
}

func (p *ImageParams) Projection() string {
	return p.projection
}

func (p *ImageParams) DataType() gdal.DataType {
	return p.datatype
}

func (p *ImageParams) NaN() (float64, bool) {
	return p.nan, p.nanPresent
}

func (p *ImageParams) NorthWest() LatLon {
	return p.Transform().Pixels2LatLon(0, 0)
}

func (p *ImageParams) SouthEast() LatLon {
	return p.Transform().Pixels2LatLon(p.XSize()-1, p.YSize()-1)
}

func (p *ImageParams) Within(ll LatLon) bool {
	x, y := p.Transform().LatLon2Pixels(ll)
	return x >= 0 && y >= 0 && x < p.XSize() && y < p.YSize()
}

func ImageParamsBuilder(xSize, ySize int) *imageParamsBuilder {
	ip := &ImageParams{
		xSize:      xSize,
		ySize:      ySize,
		transform:  AffineTransform{0, 1, 0, 0, 0, 1},
		projection: LandsatWKT,
		nan:        math.NaN(),
		nanPresent: false,
	}
	return &imageParamsBuilder{ImageParams: ip}
}

type imageParamsBuilder struct {
	*ImageParams
}

func (ipb *imageParamsBuilder) DataType(datatype gdal.DataType) *imageParamsBuilder {
	ipb.datatype = datatype
	return ipb
}


func (ipb *imageParamsBuilder) Transform(transform AffineTransform) *imageParamsBuilder {
	ipb.transform = transform
	return ipb
}

func (ipb *imageParamsBuilder) Projection(projection string) *imageParamsBuilder {
	ipb.projection = projection
	return ipb
}

func (ipb *imageParamsBuilder) NaN(nan float64) *imageParamsBuilder {
	ipb.nanPresent = true
	ipb.nan = nan
	return ipb
}

func (ipb *imageParamsBuilder) Build() *ImageParams {
	return ipb.ImageParams.copy()
}
