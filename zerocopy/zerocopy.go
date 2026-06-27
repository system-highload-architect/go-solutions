// Package zerocopy provides utilities for zero‑allocation or minimal‑allocation
// byte manipulation, JSON parsing, and buffer management. It is designed for
// high‑performance hot paths where heap allocations must be avoided.
package zerocopy

import (
	"sync"
	"unsafe"
)

// ----------------------------------------------------------------------------
// Buffer pool
// ----------------------------------------------------------------------------

var bufPool = &sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

// GetBytes returns a pointer to a pooled byte slice. The slice has zero length
// and a default capacity of 4 KB. It is intended for temporary buffers that
// will be returned to the pool via PutBytes.
//
// Example:
//
//	bufPtr := zerocopy.GetBytes()
//	buf := *bufPtr
//	defer zerocopy.PutBytes(bufPtr)
//	buf = append(buf, "data"...)
//	*bufPtr = buf
func GetBytes() *[]byte {
	return bufPool.Get().(*[]byte)
}

// PutBytes returns a byte slice pointer to the pool. The caller must not
// use the slice after this call.
func PutBytes(buf *[]byte) {
	*buf = (*buf)[:0]
	bufPool.Put(buf)
}

// ----------------------------------------------------------------------------
// Unsafe string / byte conversions (read‑only)
// ----------------------------------------------------------------------------

// StringToBytes converts a string to a byte slice without copying memory.
// WARNING: the returned slice MUST NOT be modified.
func StringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts a byte slice to a string without copying memory.
// WARNING: the original slice MUST NOT be modified while the string is in use.
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// ----------------------------------------------------------------------------
// Zero‑copy JSON helpers
// ----------------------------------------------------------------------------

// GetJSONField extracts the raw JSON value of a named field from a JSON object
// without parsing the entire document. It returns the raw bytes (including
// quotes for strings) and true if the field was found, or false otherwise.
// This function makes a single pass over the data and does not allocate.
//
// Example:
//
//	input := []byte(`{"method":"auction.bid","params":{"id":1}}`)
//	if raw, ok := zerocopy.GetJSONField(input, "method"); ok {
//	    fmt.Println(string(raw)) // "auction.bid" (with quotes)
//	}
func GetJSONField(data []byte, field string) ([]byte, bool) {
	if len(data) == 0 || len(field) == 0 {
		return nil, false
	}
	// Find the opening brace
	start := 0
	for start < len(data) && data[start] != '{' {
		start++
	}
	if start >= len(data) {
		return nil, false
	}
	// We'll do a simple scan over the JSON object
	i := start + 1
	for i < len(data) {
		// Skip whitespace
		for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
			i++
		}
		if i >= len(data) || data[i] == '}' {
			return nil, false
		}
		// Look for a string key
		if data[i] != '"' {
			return nil, false
		}
		keyStart := i
		i++
		for i < len(data) && data[i] != '"' {
			if data[i] == '\\' { // skip escaped characters
				i++
			}
			i++
		}
		if i >= len(data) {
			return nil, false
		}
		keyEnd := i
		i++ // skip closing quote
		// Compare key with field
		keyRaw := data[keyStart : keyEnd+1]
		expectedKey := `"` + field + `"`
		if len(keyRaw) == len(expectedKey) {
			matches := true
			for k := 0; k < len(keyRaw); k++ {
				if keyRaw[k] != expectedKey[k] {
					matches = false
					break
				}
			}
			if matches {
				// Skip colon and whitespace
				for i < len(data) && (data[i] == ' ' || data[i] == ':' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
					i++
				}
				if i >= len(data) {
					return nil, false
				}
				// Extract value
				valStart := i
				if data[i] == '"' {
					// string value
					i++
					for i < len(data) && data[i] != '"' {
						if data[i] == '\\' {
							i++
						}
						i++
					}
					if i < len(data) {
						i++ // include closing quote
					}
				} else if data[i] >= '0' && data[i] <= '9' || data[i] == '-' {
					// number
					i++
					for i < len(data) && (data[i] >= '0' && data[i] <= '9' || data[i] == '.' || data[i] == 'e' || data[i] == 'E' || data[i] == '+' || data[i] == '-') {
						i++
					}
				} else if data[i] == 't' || data[i] == 'f' || data[i] == 'n' {
					// true, false, null
					i += 4 // enough for all
					if i > len(data) {
						i = len(data)
					}
				} else if data[i] == '{' {
					// nested object – skip using brace counting
					i++
					depth := 1
					for i < len(data) && depth > 0 {
						if data[i] == '{' {
							depth++
						} else if data[i] == '}' {
							depth--
						}
						i++
					}
				} else if data[i] == '[' {
					// array – skip using bracket counting
					i++
					depth := 1
					for i < len(data) && depth > 0 {
						if data[i] == '[' {
							depth++
						} else if data[i] == ']' {
							depth--
						}
						i++
					}
				} else {
					return nil, false
				}
				return data[valStart:i], true
			}
		}
		// Skip value (already advanced past key, but we need to skip the value)
		// Since we didn't match, we must advance past the colon and the value
		for i < len(data) && (data[i] == ' ' || data[i] == ':' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
			i++
		}
		if i >= len(data) {
			return nil, false
		}
		// Skip value similar to above but without saving
		if data[i] == '"' {
			i++
			for i < len(data) && data[i] != '"' {
				if data[i] == '\\' {
					i++
				}
				i++
			}
			if i < len(data) {
				i++
			}
		} else if data[i] >= '0' && data[i] <= '9' || data[i] == '-' {
			i++
			for i < len(data) && (data[i] >= '0' && data[i] <= '9' || data[i] == '.' || data[i] == 'e' || data[i] == 'E' || data[i] == '+' || data[i] == '-') {
				i++
			}
		} else if data[i] == 't' || data[i] == 'f' || data[i] == 'n' {
			i += 4
		} else if data[i] == '{' {
			i++
			depth := 1
			for i < len(data) && depth > 0 {
				if data[i] == '{' {
					depth++
				} else if data[i] == '}' {
					depth--
				}
				i++
			}
		} else if data[i] == '[' {
			i++
			depth := 1
			for i < len(data) && depth > 0 {
				if data[i] == '[' {
					depth++
				} else if data[i] == ']' {
					depth--
				}
				i++
			}
		}
		// Continue to next key
	}
	return nil, false
}

// AppendJSONString appends a JSON‑escaped string to dst.
func AppendJSONString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			dst = append(dst, '\\', '"')
		case '\\':
			dst = append(dst, '\\', '\\')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, s[i])
		}
	}
	dst = append(dst, '"')
	return dst
}

// AppendJSONInt appends a decimal integer to dst.
func AppendJSONInt(dst []byte, n int64) []byte {
	// fast path for small numbers
	if n == 0 {
		return append(dst, '0')
	}
	// temporary buffer
	var tmp [20]byte
	i := len(tmp)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		i--
		tmp[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		tmp[i] = '-'
	}
	return append(dst, tmp[i:]...)
}

// ----------------------------------------------------------------------------
// Example helper (optional)
// ----------------------------------------------------------------------------
// This package is typically used in HTTP servers that need to parse JSON‑RPC
// bodies without allocating memory for each field.
