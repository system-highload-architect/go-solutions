package geospatial

import "math"

// Point представляет географическую координату в десятичных градусах.
type Point struct {
	Lat float64
	Lng float64
}

// BBox — ограничивающий прямоугольник.
type BBox struct {
	MinLat, MinLng, MaxLat, MaxLng float64
}

// bboxArea вычисляет площадь ограничивающего прямоугольника.
func bboxArea(bbox BBox) float64 {
	if bbox.MaxLat < bbox.MinLat || bbox.MaxLng < bbox.MinLng {
		return 0
	}
	return (bbox.MaxLat - bbox.MinLat) * (bbox.MaxLng - bbox.MinLng)
}

// bboxEnlargement возвращает, насколько увеличится площадь bbox при добавлении added.
func bboxEnlargement(bbox, added BBox) float64 {
	combined := combineBBox(bbox, added)
	return bboxArea(combined) - bboxArea(bbox)
}

// combineBBox объединяет два ограничивающих прямоугольника.
func combineBBox(a, b BBox) BBox {
	if a.MinLat == 0 && a.MaxLat == 0 && a.MinLng == 0 && a.MaxLng == 0 {
		return b
	}
	return BBox{
		MinLat: math.Min(a.MinLat, b.MinLat),
		MinLng: math.Min(a.MinLng, b.MinLng),
		MaxLat: math.Max(a.MaxLat, b.MaxLat),
		MaxLng: math.Max(a.MaxLng, b.MaxLng),
	}
}

// bboxIntersects возвращает true, если два ограничивающих прямоугольника пересекаются.
func bboxIntersects(a, b BBox) bool {
	return a.MinLat <= b.MaxLat && a.MaxLat >= b.MinLat &&
		a.MinLng <= b.MaxLng && a.MaxLng >= b.MinLng
}

// bboxMinDist вычисляет приближённое минимальное расстояние от точки до прямоугольника (в градусах).
func bboxMinDist(bbox BBox, point Point) float64 {
	latDist := 0.0
	if point.Lat < bbox.MinLat {
		latDist = bbox.MinLat - point.Lat
	} else if point.Lat > bbox.MaxLat {
		latDist = point.Lat - bbox.MaxLat
	}

	lngDist := 0.0
	if point.Lng < bbox.MinLng {
		lngDist = bbox.MinLng - point.Lng
	} else if point.Lng > bbox.MaxLng {
		lngDist = point.Lng - bbox.MaxLng
	}

	return math.Sqrt(latDist*latDist + lngDist*lngDist)
}

// bboxCenterDist возвращает расстояние между центрами двух BBox в метрах.
func bboxCenterDist(a, b BBox) float64 {
	ca := Point{Lat: (a.MinLat + a.MaxLat) / 2, Lng: (a.MinLng + a.MaxLng) / 2}
	cb := Point{Lat: (b.MinLat + b.MaxLat) / 2, Lng: (b.MinLng + b.MaxLng) / 2}
	return HaversineDistance(ca, cb)
}
