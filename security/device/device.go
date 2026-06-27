// Package device provides flexible, high‑performance User‑Agent parsing
// with support for user‑defined device registrations.
package device

import "strings"

// ----------------------------------------------------------------------------
// Generic classifier with hash table for exact matches and rules for partial
// ----------------------------------------------------------------------------

// Rule is a single classification rule: if the subString is found in the
// input, the value Val is assigned.
type Rule[V any] struct {
	SubString string
	Val       V
}

// Classifier matches an input string against a set of rules and an optional
// hash table for exact matches. Rules are checked in the order they were added,
// but only if no exact match is found.
type Classifier[V any] struct {
	exact map[string]V // fast O(1) lookup for exact strings
	rules []Rule[V]    // fallback for partial matches
}

// NewClassifier creates an empty classifier.
func NewClassifier[V any]() *Classifier[V] {
	return &Classifier[V]{
		exact: make(map[string]V),
	}
}

// AddExact registers an exact match for the given string.
func (c *Classifier[V]) AddExact(key string, val V) {
	c.exact[key] = val
}

// AddRule appends a partial‑match rule. Rules are evaluated in order, so more
// specific rules should be added before more general ones.
func (c *Classifier[V]) AddRule(subString string, val V) {
	c.rules = append(c.rules, Rule[V]{SubString: subString, Val: val})
}

// Classify returns the value for the input. It first checks the exact map;
// if not found, it iterates over the rules and returns the value of the
// first matching rule. If nothing matches, it returns the zero value.
func (c *Classifier[V]) Classify(input string) V {
	// 1. Try exact match (fast path)
	if val, ok := c.exact[input]; ok {
		return val
	}
	// 2. Fallback to partial rules
	for _, r := range c.rules {
		if strings.Contains(input, r.SubString) {
			return r.Val
		}
	}
	var zero V
	return zero
}

// ----------------------------------------------------------------------------
// Device types
// ----------------------------------------------------------------------------

// Type represents a broad device category.
type Type int

const (
	UnknownDevice Type = iota
	Desktop
	Mobile
	Tablet
	Bot
)

// OS represents a broad operating system category.
type OS int

const (
	UnknownOS OS = iota
	Windows
	MacOS
	Linux
	Android
	IOS
)

// Browser represents a broad browser category.
type Browser int

const (
	UnknownBrowser Browser = iota
	Chrome
	Safari
	Firefox
	Edge
	Opera
	IE
)

// DeviceInfo holds the parsed information extracted from a User‑Agent string.
type DeviceInfo struct {
	Type    Type
	OS      OS
	Browser Browser
}

// ----------------------------------------------------------------------------
// Global device registry (user‑extensible)
// ----------------------------------------------------------------------------

var (
	// deviceRegistry maps a known device token (e.g., "iPhone", "Chrome/") to
	// a DeviceInfo. Users can register additional devices before parsing.
	deviceRegistry = make(map[string]DeviceInfo)
)

// RegisterDevice adds or overrides a device entry in the global registry.
// The key is a substring that commonly appears in the User‑Agent of that device.
// The DeviceInfo will be returned by Parse if the User‑Agent contains the key.
func RegisterDevice(key string, info DeviceInfo) {
	deviceRegistry[key] = info
}

// ----------------------------------------------------------------------------
// Default classifiers (used by Parse)
// ----------------------------------------------------------------------------

var (
	defaultTypeClassifier    *Classifier[Type]
	defaultOSClassifier      *Classifier[OS]
	defaultBrowserClassifier *Classifier[Browser]
)

func init() {
	// Device types – exact matches first, then partial rules.
	defaultTypeClassifier = NewClassifier[Type]()
	defaultTypeClassifier.AddExact("Tablet", Tablet)
	defaultTypeClassifier.AddExact("iPad", Tablet)
	defaultTypeClassifier.AddExact("PlayBook", Tablet)
	defaultTypeClassifier.AddRule("Mobi", Mobile)
	defaultTypeClassifier.AddRule("Android", Mobile) // Android without Tablet already handled
	defaultTypeClassifier.AddRule("Googlebot", Bot)
	defaultTypeClassifier.AddRule("Bingbot", Bot)
	defaultTypeClassifier.AddRule("Slurp", Bot)
	defaultTypeClassifier.AddRule("DuckDuckBot", Bot)
	defaultTypeClassifier.AddRule("Windows NT", Desktop)
	defaultTypeClassifier.AddRule("Macintosh", Desktop)
	defaultTypeClassifier.AddRule("X11", Desktop)

	// Operating systems – exact matches first.
	defaultOSClassifier = NewClassifier[OS]()
	defaultOSClassifier.AddExact("Windows NT", Windows)
	defaultOSClassifier.AddExact("Macintosh", MacOS)
	defaultOSClassifier.AddExact("Mac OS X", MacOS)
	defaultOSClassifier.AddExact("Android", Android)
	defaultOSClassifier.AddExact("iPhone", IOS)
	defaultOSClassifier.AddExact("iPad", IOS)
	defaultOSClassifier.AddExact("iPod", IOS)
	defaultOSClassifier.AddExact("Linux", Linux)

	// Browsers – exact matches first (order matters because Chrome UA contains Safari).
	defaultBrowserClassifier = NewClassifier[Browser]()
	defaultBrowserClassifier.AddExact("Edge/", Edge)
	defaultBrowserClassifier.AddExact("Chrome/", Chrome)
	defaultBrowserClassifier.AddExact("Safari/", Safari)
	defaultBrowserClassifier.AddExact("Firefox/", Firefox)
	defaultBrowserClassifier.AddExact("OPR/", Opera)
	defaultBrowserClassifier.AddExact("Opera", Opera)
	defaultBrowserClassifier.AddExact("MSIE", IE)
	defaultBrowserClassifier.AddExact("Trident", IE)
}

// Parse extracts device type, operating system, and browser from a raw
// User‑Agent string. It first checks the user‑defined device registry;
// if a key is found, its DeviceInfo is returned immediately. Otherwise
// the default classifiers are used.
func Parse(ua string) DeviceInfo {
	// Check user‑defined registry first (fast O(1) map lookup).
	for key, info := range deviceRegistry {
		if strings.Contains(ua, key) {
			return info
		}
	}
	// Fallback to default classifiers.
	return DeviceInfo{
		Type:    defaultTypeClassifier.Classify(ua),
		OS:      defaultOSClassifier.Classify(ua),
		Browser: defaultBrowserClassifier.Classify(ua),
	}
}
