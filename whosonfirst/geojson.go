package geojson

import (
	"github.com/jeffail/gabs"
	"github.com/dhconnelly/rtreego"
	"io/ioutil"
	"fmt"
)

/*
Something like this using "github.com/paulmach/go.geojson" seems
like it would be a good thing but I don't think I have the stamina
to figure out how to parse the properties separately right now...
(20151005/thisisaaronland)
/*

type WOFProperties struct {
     Raw    []byte
     Parsed *gabs.Container
}

type WOFFeature struct {
     ID          json.Number            `json:"id,omitempty"`
     Type        string                 `json:"type"`
     BoundingBox []float64              `json:"bbox,omitempty"`	// maybe make this a WOFBounds (rtree) like properties?
     Geometry    *gj.Geometry      `json:"geometry"`
     Properties  WOFProperties			`json:"properties"`
     // Properties  map[string]interface{} `json:"properties"`
     CRS         map[string]interface{} `json:"crs,omitempty"` // Coordinate Reference System Objects are not currently supported
}
*/

type WOFFeature struct {
	Raw    []byte
	Parsed *gabs.Container
}

func (wof WOFFeature) Body() *gabs.Container {
	return wof.Parsed
}

func (wof WOFFeature) Dumps() string {
	return wof.Parsed.String()
}

// See notes above in WOFFeature.BoundingBox - for now this will do...
// (20151012/thisisaaronland)

func (wof WOFFeature) Bounds() (*rtreego.Rect, error) {

	body := wof.Body()
	children, _ := body.S("bbox").Children()

	var swlon float64
	var swlat float64
	var nelon float64
	var nelat float64

	swlon = children[0].Data().(float64)
	swlat = children[1].Data().(float64)
	nelon = children[2].Data().(float64)
	nelat = children[3].Data().(float64)

	llat := nelat - swlat
	llon := nelon - swlon

	// fmt.Printf("%f - %f = %f\n", nelat, swlat, llat)
	// fmt.Printf("%f - %f = %f\n", nelon, swlon, llon)

	pt := rtreego.Point{swlon, swlat}
	rect, err := rtreego.NewRect(pt, []float64{llon, llat})

     	if err != nil {
     	   return nil, err
     	}

     	return rect, nil
}

func UnmarshalFile(path string) (*WOFFeature, error) {

	body, read_err := ioutil.ReadFile(path)

	if read_err != nil {
		return nil, read_err
	}

	return UnmarshalFeature(body)
}

func UnmarshalFeature(raw []byte) (*WOFFeature, error) {

	parsed, parse_err := gabs.ParseJSON(raw)

	if parse_err != nil {
		return nil, parse_err
	}

	rsp := WOFFeature{
		Raw:    raw,
		Parsed: parsed,
	}

	return &rsp, nil
}