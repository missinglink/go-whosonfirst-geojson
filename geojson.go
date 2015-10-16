package geojson

import (
	rtreego "github.com/dhconnelly/rtreego"
	gabs "github.com/jeffail/gabs"
	geo "github.com/kellydunn/golang-geo"
	ioutil "io/ioutil"
)

/*

- gabs is what handles marshaling a random bag of GeoJSON
- rtreego is imported to convert a WOFFeature in to a handy rtreego.Spatial object for indexing by go-whosonfirst-pip
- geo is imported to convert a WOFFeature geometry into a list of geo.Polygon objects for doing containment checks in go-whosonfirst-pip
  (only Polygons and MultiPolygons are supported at the moment)

*/

type WOFError struct {
	s string
}

func (e *WOFError) Error() string {
	return e.s
}

// See also
// https://github.com/dhconnelly/rtreego#storing-updating-and-deleting-objects

type WOFSpatial struct {
	bounds    *rtreego.Rect
	Id        int
	Name      string
	Placetype string
}

func (sp WOFSpatial) Bounds() *rtreego.Rect {
	return sp.bounds
}

type WOFFeature struct {
	// Raw    []byte
	Parsed *gabs.Container
}

func (wof WOFFeature) Body() *gabs.Container {
	return wof.Parsed
}

func (wof WOFFeature) Dumps() string {
	return wof.Parsed.String()
}

func (wof WOFFeature) WOFId() int {

	body := wof.Body()

	var flid float64
	var id int

	var ok bool

	// what follows shouldn't be necessary but appears to be
	// for... uh, reasons (20151013/thisisaaronland)

	flid, ok = body.Path("properties.wof:id").Data().(float64)

	if ok {
		id = int(flid)
	} else {
		id, ok = body.Path("properties.wof:id").Data().(int)
	}

	if !ok {
		id = -1
	}

	return id
}

func (wof WOFFeature) WOFName() string {

	body := wof.Body()

	var name string
	var ok bool

	name, ok = body.Path("properties.wof:name").Data().(string)

	if !ok {
		name = ""
	}

	return name
}

// Should return a full-on WOFPlacetype object thing-y
// (20151012/thisisaaronland)

func (wof WOFFeature) WOFPlacetype() string {

	body := wof.Body()

	var placetype string
	var ok bool

	placetype, ok = body.Path("properties.wof:placetype").Data().(string)

	if !ok {
		placetype = "unknown"
	}

	return placetype
}

func (wof WOFFeature) EnSpatialize() (*WOFSpatial, error) {

	id := wof.WOFId()
	name := wof.WOFName()
	placetype := wof.WOFPlacetype()

	body := wof.Body()

	var swlon float64
	var swlat float64
	var nelon float64
	var nelat float64

	children, _ := body.S("bbox").Children()

	if len(children) != 4 {
		return nil, &WOFError{"weird and freaky bounding box"}
	}

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

	return &WOFSpatial{rect, id, name, placetype}, nil
}

func (wof WOFFeature) GeomToPolygons() []*geo.Polygon {

	body := wof.Body()

	var geom_type string

	geom_type, _ = body.Path("geometry.type").Data().(string)
	children, _ := body.S("geometry").ChildrenMap()

	polygons := make([]*geo.Polygon, 0)

	for key, child := range children {

		if key != "coordinates" {
			continue
		}

		var coordinates []interface{}
		coordinates, _ = child.Data().([]interface{})

		if geom_type == "Polygon" {
			polygons = wof.DumpPolygon(coordinates)
		} else if geom_type == "MultiPolygon" {
			polygons = wof.DumpMultiPolygon(coordinates)
		} else {
			// pass
		}
	}

	return polygons
}

func (wof WOFFeature) DumpMultiPolygon(coordinates []interface{}) []*geo.Polygon {

	polygons := make([]*geo.Polygon, 0)

	for _, ipolys := range coordinates {

		polys := ipolys.([]interface{})

		for _, ipoly := range polys {

			poly := ipoly.([]interface{})
			polygon := wof.DumpCoords(poly)
			polygons = append(polygons, polygon)
		}

	}

	return polygons
}

func (wof WOFFeature) DumpPolygon(coordinates []interface{}) []*geo.Polygon {

	polygons := make([]*geo.Polygon, 0)

	for _, ipoly := range coordinates {

		poly := ipoly.([]interface{})
		polygon := wof.DumpCoords(poly)
		polygons = append(polygons, polygon)
	}

	return polygons
}

func (wof WOFFeature) DumpCoords(poly []interface{}) *geo.Polygon {

	polygon := &geo.Polygon{}

	for _, icoords := range poly {

		coords := icoords.([]interface{})

		lon := coords[0].(float64)
		lat := coords[1].(float64)

		pt := geo.NewPoint(lat, lon)
		polygon.Add(pt)
	}

	return polygon
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
		// Raw:    raw,
		Parsed: parsed,
	}

	return &rsp, nil
}
