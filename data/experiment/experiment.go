// Package experiment provides deterministic A/B experiment assignment.
package experiment

import (
	"hash/fnv"
)

// Experiments holds named experiment flags with their respective traffic percentages.
type Experiments struct {
	flags map[string]float64 // name -> percentage (0.0–1.0)
}

// New creates a new Experiments instance.
func New(flags map[string]float64) *Experiments {
	if flags == nil {
		flags = make(map[string]float64)
	}
	return &Experiments{flags: flags}
}

// IsInExperiment returns true if the given userID falls into the specified experiment.
// The decision is deterministic based on the hash of userID + experiment name.
func (e *Experiments) IsInExperiment(userID, name string) bool {
	p, ok := e.flags[name]
	if !ok || p <= 0 {
		return false
	}
	if p >= 1 {
		return true
	}
	h := fnv.New64a()
	h.Write([]byte(userID))
	h.Write([]byte(name))
	hash := h.Sum64()
	return float64(hash%10000)/10000.0 < p
}
