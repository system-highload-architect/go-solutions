// Package radixsort provides LSD (least significant digit) Radix Sort implementations
// for integer and floating-point types. All functions sort in ascending order.
// Time complexity: O(N) for N elements. Space complexity: O(N) temporary buffer.
package radixsort

import (
	"math"
)

// SortInt64 sorts a slice of int64 in ascending order using LSD Radix Sort.
// It performs 8 passes (8 bytes) and uses a temporary buffer of the same length.
func SortInt64(data []int64) {
	n := len(data)
	if n < 2 {
		return
	}
	// convert to unsigned: flip sign bit
	src := make([]uint64, n)
	for i, v := range data {
		src[i] = uint64(v) ^ (1 << 63)
	}
	dst := make([]uint64, n)

	// 8 passes, 8 bits per pass
	for shift := 0; shift < 64; shift += 8 {
		var count [256]int
		for _, v := range src {
			b := (v >> shift) & 0xFF
			count[b]++
		}
		// prefix sums
		for i := 1; i < 256; i++ {
			count[i] += count[i-1]
		}
		// build output (stable, iterate in reverse)
		for i := n - 1; i >= 0; i-- {
			b := (src[i] >> shift) & 0xFF
			count[b]--
			dst[count[b]] = src[i]
		}
		src, dst = dst, src
	}
	// copy back and convert to int64
	for i, v := range src {
		data[i] = int64(v ^ (1 << 63))
	}
}

// SortInt64WithIndices sorts data and simultaneously reorders indices.
// After the call, data is sorted and indices[i] is the original position of data[i].
func SortInt64WithIndices(data []int64, indices []int) {
	n := len(data)
	if n < 2 || len(indices) != n {
		return
	}
	type pair struct {
		val uint64
		idx int
	}
	src := make([]pair, n)
	for i, v := range data {
		src[i] = pair{uint64(v) ^ (1 << 63), i}
	}
	dst := make([]pair, n)

	for shift := 0; shift < 64; shift += 8 {
		var count [256]int
		for _, p := range src {
			b := (p.val >> shift) & 0xFF
			count[b]++
		}
		for i := 1; i < 256; i++ {
			count[i] += count[i-1]
		}
		for i := n - 1; i >= 0; i-- {
			b := (src[i].val >> shift) & 0xFF
			count[b]--
			dst[count[b]] = src[i]
		}
		src, dst = dst, src
	}
	for i, p := range src {
		data[i] = int64(p.val ^ (1 << 63))
		indices[i] = p.idx
	}
}

// SortUint64 sorts a slice of uint64 in ascending order.
func SortUint64(data []uint64) {
	n := len(data)
	if n < 2 {
		return
	}
	src := make([]uint64, n)
	copy(src, data)
	dst := make([]uint64, n)

	for shift := 0; shift < 64; shift += 8 {
		var count [256]int
		for _, v := range src {
			b := (v >> shift) & 0xFF
			count[b]++
		}
		for i := 1; i < 256; i++ {
			count[i] += count[i-1]
		}
		for i := n - 1; i >= 0; i-- {
			b := (src[i] >> shift) & 0xFF
			count[b]--
			dst[count[b]] = src[i]
		}
		src, dst = dst, src
	}
	copy(data, src)
}

// SortFloat64 sorts a slice of float64 in ascending order.
// It converts floats to their integer representation such that the ordering is preserved.
func SortFloat64(data []float64) {
	n := len(data)
	if n < 2 {
		return
	}
	src := make([]uint64, n)
	for i, v := range data {
		bits := math.Float64bits(v)
		// If sign bit is 1 (negative), flip all bits; otherwise, flip only sign bit.
		if bits>>63 == 1 {
			bits = ^bits
		} else {
			bits ^= 1 << 63
		}
		src[i] = bits
	}
	dst := make([]uint64, n)

	for shift := 0; shift < 64; shift += 8 {
		var count [256]int
		for _, v := range src {
			b := (v >> shift) & 0xFF
			count[b]++
		}
		for i := 1; i < 256; i++ {
			count[i] += count[i-1]
		}
		for i := n - 1; i >= 0; i-- {
			b := (src[i] >> shift) & 0xFF
			count[b]--
			dst[count[b]] = src[i]
		}
		src, dst = dst, src
	}
	for i, bits := range src {
		// convert back
		if bits>>63 == 1 {
			bits ^= 1 << 63
		} else {
			bits = ^bits
		}
		data[i] = math.Float64frombits(bits)
	}
}
