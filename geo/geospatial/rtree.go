package geospatial

import (
	"math"
	"sort"
)

// RTree — двумерное R‑дерево для пространственного индексирования.
type RTree[T any] struct {
	root       *rNode[T]
	maxEntries int
	minEntries int
	size       int
}

type rNode[T any] struct {
	children []*rNode[T]
	entries  []rEntry[T]
	bbox     BBox
	isLeaf   bool
}

type rEntry[T any] struct {
	bbox  BBox
	value T
}

// distEntry используется внутри Nearest.
type distEntry[T any] struct {
	value T
	dist  float64
}

// NewRTree создаёт пустое R‑дерево.
func NewRTree[T any](maxEntries int) *RTree[T] {
	if maxEntries < 4 {
		maxEntries = 4
	}
	return &RTree[T]{
		maxEntries: maxEntries,
		minEntries: maxEntries / 2,
		root: &rNode[T]{
			isLeaf:   true,
			children: make([]*rNode[T], 0, maxEntries),
			entries:  make([]rEntry[T], 0, maxEntries),
		},
	}
}

// Insert добавляет объект в дерево.
func (rt *RTree[T]) Insert(bbox BBox, value T) {
	leaf := rt.chooseLeaf(rt.root, bbox)
	leaf.entries = append(leaf.entries, rEntry[T]{bbox: bbox, value: value})
	leaf.bbox = combineBBox(leaf.bbox, bbox)
	rt.size++

	if len(leaf.entries) > rt.maxEntries {
		rt.split(leaf)
	}
}

// Search возвращает все объекты, чьи bounding box пересекаются с query.
func (rt *RTree[T]) Search(query BBox) []T {
	var result []T
	rt.search(rt.root, query, &result)
	return result
}

// Nearest возвращает до k ближайших соседей к точке point.
func (rt *RTree[T]) Nearest(point Point, k int) []T {
	var candidates []distEntry[T]
	rt.nearest(rt.root, point, k, &candidates)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist < candidates[j].dist
	})
	if len(candidates) > k {
		candidates = candidates[:k]
	}
	result := make([]T, len(candidates))
	for i, c := range candidates {
		result[i] = c.value
	}
	return result
}

// Size возвращает количество объектов в дереве.
func (rt *RTree[T]) Size() int {
	return rt.size
}

// Внутренние методы

func (rt *RTree[T]) chooseLeaf(node *rNode[T], bbox BBox) *rNode[T] {
	if node.isLeaf {
		return node
	}
	var bestChild *rNode[T]
	var bestEnlargement float64 = math.MaxFloat64
	var bestArea float64

	for _, child := range node.children {
		enlargement := bboxEnlargement(child.bbox, bbox)
		area := bboxArea(child.bbox)
		if bestChild == nil || enlargement < bestEnlargement || (enlargement == bestEnlargement && area < bestArea) {
			bestChild = child
			bestEnlargement = enlargement
			bestArea = area
		}
	}
	return rt.chooseLeaf(bestChild, bbox)
}

func (rt *RTree[T]) search(node *rNode[T], query BBox, result *[]T) {
	if !bboxIntersects(node.bbox, query) {
		return
	}
	if node.isLeaf {
		for _, entry := range node.entries {
			if bboxIntersects(entry.bbox, query) {
				*result = append(*result, entry.value)
			}
		}
		return
	}
	for _, child := range node.children {
		rt.search(child, query, result)
	}
}

func (rt *RTree[T]) nearest(node *rNode[T], point Point, k int, candidates *[]distEntry[T]) {
	if node.isLeaf {
		for _, entry := range node.entries {
			dist := bboxMinDist(entry.bbox, point)
			*candidates = append(*candidates, distEntry[T]{value: entry.value, dist: dist})
		}
		return
	}
	type childDist[T any] struct {
		node *rNode[T]
		dist float64
	}
	var children []childDist[T]
	for _, child := range node.children {
		children = append(children, childDist[T]{node: child, dist: bboxMinDist(child.bbox, point)})
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].dist < children[j].dist
	})
	for _, cd := range children {
		if len(*candidates) >= k && cd.dist > (*candidates)[k-1].dist {
			continue
		}
		rt.nearest(cd.node, point, k, candidates)
	}
}

func (rt *RTree[T]) split(node *rNode[T]) {
	if len(node.entries) <= 1 {
		return
	}
	seed1, seed2 := 0, 1
	maxDist := 0.0
	for i := 0; i < len(node.entries); i++ {
		for j := i + 1; j < len(node.entries); j++ {
			d := bboxCenterDist(node.entries[i].bbox, node.entries[j].bbox)
			if d > maxDist {
				maxDist = d
				seed1, seed2 = i, j
			}
		}
	}

	left := &rNode[T]{isLeaf: node.isLeaf}
	right := &rNode[T]{isLeaf: node.isLeaf}
	left.entries = append(left.entries, node.entries[seed1])
	right.entries = append(right.entries, node.entries[seed2])
	left.bbox = node.entries[seed1].bbox
	right.bbox = node.entries[seed2].bbox

	for i, entry := range node.entries {
		if i == seed1 || i == seed2 {
			continue
		}
		enlargeLeft := bboxEnlargement(left.bbox, entry.bbox)
		enlargeRight := bboxEnlargement(right.bbox, entry.bbox)
		if enlargeLeft < enlargeRight {
			left.entries = append(left.entries, entry)
			left.bbox = combineBBox(left.bbox, entry.bbox)
		} else {
			right.entries = append(right.entries, entry)
			right.bbox = combineBBox(right.bbox, entry.bbox)
		}
	}

	if node == rt.root {
		newRoot := &rNode[T]{
			isLeaf:   false,
			children: []*rNode[T]{left, right},
			bbox:     combineBBox(left.bbox, right.bbox),
		}
		rt.root = newRoot
	} else {
		parent := rt.findParent(rt.root, node)
		if parent != nil {
			parent.children = removeChild[T](parent.children, node)
			parent.children = append(parent.children, left, right)
			parent.bbox = combineBBox(parent.bbox, left.bbox)
			parent.bbox = combineBBox(parent.bbox, right.bbox)
			if len(parent.children) > rt.maxEntries {
				rt.split(parent)
			}
		}
	}
}

func (rt *RTree[T]) findParent(node, target *rNode[T]) *rNode[T] {
	if node.isLeaf {
		return nil
	}
	for _, child := range node.children {
		if child == target {
			return node
		}
		if p := rt.findParent(child, target); p != nil {
			return p
		}
	}
	return nil
}

func removeChild[T any](children []*rNode[T], target *rNode[T]) []*rNode[T] {
	for i, c := range children {
		if c == target {
			return append(children[:i], children[i+1:]...)
		}
	}
	return children
}
