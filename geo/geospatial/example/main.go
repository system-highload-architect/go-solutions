// Пример использования R‑дерева.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/geo/geospatial"
)

type Driver struct {
	ID    int
	Name  string
	Point geospatial.Point
}

func main() {
	rt := geospatial.NewRTree[*Driver](16)

	drivers := []Driver{
		{ID: 1, Name: "Алексей", Point: geospatial.Point{Lat: 55.7558, Lng: 37.6173}},
		{ID: 2, Name: "Борис", Point: geospatial.Point{Lat: 55.7512, Lng: 37.6184}},
		{ID: 3, Name: "Виктор", Point: geospatial.Point{Lat: 55.7600, Lng: 37.6200}},
	}
	for i := range drivers {
		d := &drivers[i]
		rt.Insert(geospatial.BBox{
			MinLat: d.Point.Lat, MaxLat: d.Point.Lat,
			MinLng: d.Point.Lng, MaxLng: d.Point.Lng,
		}, d)
	}

	client := geospatial.Point{Lat: 55.7530, Lng: 37.6190}
	nearest := rt.Nearest(client, 2)
	for _, d := range nearest {
		fmt.Printf("Водитель %s (ID %d) на расстоянии %.0f м\n",
			d.Name, d.ID, geospatial.HaversineDistance(client, d.Point))
	}
}
