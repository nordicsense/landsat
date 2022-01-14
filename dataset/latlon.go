package dataset

import (
	"errors"
	"fmt"
	"math"

	"github.com/nordicsense/gdal"
)

const LandsatWKT = `PROJCRS["WGS 84 / UTM zone 36N",
    BASEGEOGCRS["WGS 84",
        DATUM["World Geodetic System 1984",
            ELLIPSOID["WGS 84",6378137,298.257223563,
                LENGTHUNIT["metre",1]]],
        PRIMEM["Greenwich",0,
            ANGLEUNIT["degree",0.0174532925199433]],
        ID["EPSG",4326]],
    CONVERSION["UTM zone 36N",
        METHOD["Transverse Mercator",
            ID["EPSG",9807]],
        PARAMETER["Latitude of natural origin",0,
            ANGLEUNIT["degree",0.0174532925199433],
            ID["EPSG",8801]],
        PARAMETER["Longitude of natural origin",33,
            ANGLEUNIT["degree",0.0174532925199433],
            ID["EPSG",8802]],
        PARAMETER["Scale factor at natural origin",0.9996,
            SCALEUNIT["unity",1],
            ID["EPSG",8805]],
        PARAMETER["False easting",500000,
            LENGTHUNIT["metre",1],
            ID["EPSG",8806]],
        PARAMETER["False northing",0,
            LENGTHUNIT["metre",1],
            ID["EPSG",8807]]],
    CS[Cartesian,2],
        AXIS["(E)",east,
            ORDER[1],
            LENGTHUNIT["metre",1]],
        AXIS["(N)",north,
            ORDER[2],
            LENGTHUNIT["metre",1]],
    USAGE[
        SCOPE["Engineering survey, topographic mapping."],
        AREA["Between 30°E and 36°E, northern hemisphere between equator and 84°N, onshore and offshore. Belarus. Cyprus. Egypt. Ethiopia. Finland. Israel. Jordan. Kenya. Lebanon. Moldova. Norway. Russian Federation. Saudi Arabia. Sudan. Syria. Turkey. Uganda. Ukraine."],
        BBOX[0,30,84,36]],
    ID["EPSG",32636]]`

// LatLon represents a latitude/longitude pair.
type LatLon [2]float64

// Transform coordinates from one ESPG projection into another.
func (ll LatLon) Transform(fromESPG, toESPG int) (LatLon, error) {
	from, err := ll.CSRFromESPG(fromESPG)
	if err != nil {
		return ll, err
	}
	defer from.Destroy()
	to, err := ll.CSRFromESPG(toESPG)
	if err != nil {
		return ll, err
	}
	defer to.Destroy()
	return ll.transform(from, to)
}

func (ll LatLon) transform(from, to gdal.SpatialReference) (LatLon, error) {
	t := gdal.CreateCoordinateTransform(from, to)
	defer t.Destroy()
	lat := []float64{ll[0]}
	lon := []float64{ll[1]}
	z := []float64{0.0}
	if ok := t.Transform(1, lon, lat, z); ok {
		return LatLon{lat[0], lon[0]}, nil
	}
	return LatLon{lat[0], lon[0]}, errors.New("transformation failed")
}

func (ll LatLon) CSRFromESPG(espg int) (gdal.SpatialReference, error) {
	res := gdal.CreateSpatialReference("")
	err := res.FromEPSG(espg)
	return res, err
}

func (ll LatLon) LANDSAT_CSR() gdal.SpatialReference {
	return gdal.CreateSpatialReference(LandsatWKT)
}

// Degrees2Sin transforms coordinates from the World Geodetic System (WGS84, given in degrees) into Sphere Sinusoidal.
func (ll LatLon) Degrees2Sin() (LatLon, error) {
	from, err := ll.CSRFromESPG(4326)
	if err != nil {
		return ll, err
	}
	to := ll.LANDSAT_CSR()
	defer from.Destroy()
	defer to.Destroy()
	return ll.transform(from, to)
}

// Sin2Degree transforms coordinates from the Sphere Sinusoidal system into the World Geodetic System (WGS84).
func (ll LatLon) Sin2Degree() (LatLon, error) {
	from := ll.LANDSAT_CSR()
	defer from.Destroy()
	to, err := ll.CSRFromESPG(4326)
	if err != nil {
		return ll, err
	}
	defer to.Destroy()
	return ll.transform(from, to)
}

func (ll LatLon) String() string {
	return fmt.Sprintf("(%.2f,%.2f)", ll[0], ll[1])
}

// AffineTransform defines the transformation of the projection.
type AffineTransform [6]float64

// Pixels2LatLonSin performs the direct affine transform from image pixels to World Sinusoidal coordinates.
func (at AffineTransform) Pixels2LatLonSin(x, y int) LatLon {
	lat := float64(y)*at[5] + at[3]
	lon := float64(x)*at[1] + at[0]
	return LatLon{lat, lon}
}

// Pixels2LatLon performs the direct affine transform from image pixels to lat/lon in degrees.
func (at AffineTransform) Pixels2LatLon(x, y int) LatLon {
	lat := float64(y)*at[5] + at[3]
	lon := float64(x)*at[1] + at[0]
	res, _ := LatLon{lat, lon}.Sin2Degree()
	return res
}

// LatLonSin2Pixels performs the inverse affine transform from World Sinusoidal coordinates to image pixels.
func (at AffineTransform) LatLonSin2Pixels(ll LatLon) (int, int) {
	x := int(math.Round((ll[1] - at[0]) / at[1]))
	y := int(math.Round((ll[0] - at[3]) / at[5]))
	return x, y
}

// LatLon2Pixels performs the inverse affine transform from lat/lon in degrees to image pixels.
func (at AffineTransform) LatLon2Pixels(ll LatLon) (int, int) {
	ll, _ = ll.Degrees2Sin()
	return at.LatLonSin2Pixels(ll)
}
