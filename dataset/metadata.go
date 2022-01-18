package dataset

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strconv"
	"time"
)

type ImageMetadata struct {
	Date         time.Time
	SunElevation float64
	SunAzimuth   float64
	Aux          map[string]string
	Bands        map[int]BandMetadata
}

type BandMetadata struct {
	RefScale, RefOffset float64
	RefMin, RefMax      float64
	RadScale, RadOffset float64
	RadMin, RadMax      float64
}

func ParseMetadata(root, prefix string) (ImageMetadata, error) {
	var (
		err   error
		jf    *os.File
		bytes []byte
		data  map[string]interface{}
	)

	fi := path.Join(root, prefix+"_MTL.json")
	if jf, err = os.Open(fi); err == nil {
		defer func() { _ = jf.Close() }()
		if bytes, err = ioutil.ReadAll(jf); err == nil {
			err = json.Unmarshal(bytes, &data)
		}
	}
	im := ImageMetadata{Aux: make(map[string]string), Bands: make(map[int]BandMetadata)}
	if err != nil {
		return im, err
	}
	data = (data["LANDSAT_METADATA_FILE"]).(map[string]interface{})

	image := (data["IMAGE_ATTRIBUTES"]).(map[string]interface{})
	if im.Date, err = time.Parse("2006-01-02", (image["DATE_ACQUIRED"]).(string)); err != nil {
		return im, err
	}
	if im.SunElevation, err = strconv.ParseFloat((image["SUN_ELEVATION"]).(string), 64); err != nil {
		return im, err
	}
	if im.SunAzimuth, err = strconv.ParseFloat((image["SUN_AZIMUTH"]).(string), 64); err != nil {
		return im, err
	}
	for _, k := range []string{"SPACECRAFT_ID", "WRS_TYPE", "WRS_PATH", "WRS_ROW", "CLOUD_COVER", "CLOUD_COVER_LAND"} {
		im.Aux[k] = (image[k]).(string)
	}

	get := func(data map[string]interface{}, k string, def float64) float64 {
		smth, ok := data[k]
		if !ok {
			return def
		}
		if val, err := strconv.ParseFloat(smth.(string), 64); err != nil {
			return def
		} else {
			return val
		}
	}

	scaling := (data["LEVEL1_RADIOMETRIC_RESCALING"]).(map[string]interface{})
	refMinmax := (data["LEVEL1_MIN_MAX_REFLECTANCE"]).(map[string]interface{})
	radMinMax := (data["LEVEL1_MIN_MAX_RADIANCE"]).(map[string]interface{})
	for i := 1; i <= 7; i++ {
		bm := BandMetadata{}
		suffix := "_BAND_" + strconv.Itoa(i)
		bm.RadMin = get(radMinMax, "RADIANCE_MINIMUM"+suffix, math.NaN())
		bm.RadMax = get(radMinMax, "RADIANCE_MAXIMUM"+suffix, math.NaN())
		bm.RefMin = get(refMinmax, "REFLECTANCE_MINIMUM"+suffix, math.NaN())
		bm.RefMax = get(refMinmax, "REFLECTANCE_MAXIMUM"+suffix, math.NaN())
		bm.RadScale = get(scaling, "RADIANCE_MULT"+suffix, 1.0)
		bm.RadOffset = get(scaling, "RADIANCE_ADD"+suffix, 0.0)
		bm.RefScale = get(scaling, "REFLECTANCE_MULT"+suffix, 1.0)
		bm.RefOffset = get(scaling, "REFLECTANCE_ADD"+suffix, 0.0)
		im.Bands[i] = bm
	}
	return im, nil
}
