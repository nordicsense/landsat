package dataset

type RasterParams struct {
	offset     float64
	scale      float64
	metadata   map[string]string
}

func (p *RasterParams) copy() *RasterParams {
	res := &RasterParams{
		offset:     p.offset,
		scale:      p.scale,
		metadata:   make(map[string]string),
	}
	for k, v := range p.metadata {
		res.metadata[k] = v
	}
	return res
}

func (p *RasterParams) ToBuilder() *rasterParamsBuilder {
	return &rasterParamsBuilder{RasterParams: p.copy()}
}

func (p *RasterParams) Offset() float64 {
	return p.offset
}

func (p *RasterParams) Scale() float64 {
	return p.scale
}

func (p *RasterParams) Metadata() map[string]string {
	return p.metadata // TODO: copy or protect
}

func RasterParamsBuilder() *rasterParamsBuilder {
	ip := &RasterParams{
		offset:     0.0,
		scale:      1.0,
		metadata:   make(map[string]string),
	}
	return &rasterParamsBuilder{RasterParams: ip}
}

type rasterParamsBuilder struct {
	*RasterParams
}

func (ipb *rasterParamsBuilder) Offset(offset float64) *rasterParamsBuilder {
	ipb.offset = offset
	return ipb
}

func (ipb *rasterParamsBuilder) Scale(scale float64) *rasterParamsBuilder {
	ipb.scale = scale
	return ipb
}

func (ipb *rasterParamsBuilder) Metadata(key, value string) *rasterParamsBuilder {
	ipb.metadata[key] = value
	return ipb
}

func (ipb *rasterParamsBuilder) Build() *RasterParams {
	return ipb.RasterParams.copy()
}
