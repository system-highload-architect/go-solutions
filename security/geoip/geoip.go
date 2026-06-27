// Package geoip provides in‑memory GeoIP lookups using MaxMind GeoLite2 databases.
package geoip

import (
	"net"
	"os"

	"github.com/oschwald/geoip2-golang"
)

// Result holds the geographic information obtained from a GeoIP lookup.
type Result struct {
	Country string
	City    string
	Lat     float64
	Lng     float64
}

// GeoDB wraps a MaxMind GeoLite2 reader and provides a safe Lookup method.
type GeoDB struct {
	reader *geoip2.Reader
}

// New opens a MaxMind GeoLite2 database file and returns a ready‑to‑use GeoDB.
func New(path string) (*GeoDB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	reader, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &GeoDB{reader: reader}, nil
}

// Lookup resolves an IP address string and returns geographic information.
// The IP string must be a valid IPv4 or IPv6 address.
func (g *GeoDB) Lookup(ipStr string) (Result, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Result{}, &net.ParseError{Type: "IP address", Text: ipStr}
	}

	record, err := g.reader.City(ip)
	if err != nil {
		return Result{}, err
	}

	res := Result{
		Country: record.Country.Names["en"],
		Lat:     record.Location.Latitude,
		Lng:     record.Location.Longitude,
	}
	if len(record.City.Names) > 0 {
		res.City = record.City.Names["en"]
	}
	return res, nil
}

// Close releases resources held by the underlying MaxMind reader.
func (g *GeoDB) Close() error {
	return g.reader.Close()
}
