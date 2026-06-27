package geospatial

import "math"

// HaversineDistance возвращает расстояние по большому кругу в метрах между двумя точками.
func HaversineDistance(a, b Point) float64 {
	const R = 6371000 // радиус Земли в метрах
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180

	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)

	aHav := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLng*sinDLng
	return 2 * R * math.Atan2(math.Sqrt(aHav), math.Sqrt(1-aHav))
}
