// LatLng<->GeoHash conversion
// Based on the ubiquitous https://github.com/davetroy/geohash-js
//
// For more information see: http://en.wikipedia.org/wiki/GeohashA
// And for visualizing: http://geohash.2ch.to/,
//  http://www.bigdatamodeling.org/2013/01/intuitive-geohash.html
//
// The quick takeaway:
// - truncating a hash reduces the precision (geohashes implicitly
//   represent non-uniformly sized cells on the earth, the longer the
//   code the smaller the area of the cell).
// - if two geohashes share a prefix, they are "near by". The longer
//   the shared prefix, the physically closer they are.
// - the converse is not always true.  Two close-by points are not
//   guaranteed to share a long prefix (consider neighboring cells
//   when searching for nearby points).
//
// This is the geohash grid (each base-32 digit subdivides space
// using this pattern).  It's a z-ordered tiling of 8 4x4 blocks,
// starting in the lower left.
//  +-------+-------+-------+-------+
//  | B | C | F | G | U | V | Y | Z |
//  |-------|-------|-------|-------|
//  | 8 | 9 | D | E | S | T | W | X |
//  +-------+-------+-------+-------+
//  | 2 | 3 | 6 | 7 | K | M | Q | R |
//  |-------|-------|-------|-------|
//  | 0 | 1 | 4 | 5 | H | J | N | P |
//  +-------+-------+-------+-------+
//

package geohash

import (
	"bytes"
	"fmt"
)

type GeoHash []byte
type Direction int

const (
	Dir_UP    Direction = 0
	Dir_RIGHT Direction = 1
	Dir_DOWN  Direction = 2
	Dir_LEFT  Direction = 3
)

var (
	// base32 digits used by GeoHash
	base32 = []byte("0123456789bcdefghjkmnpqrstuvwxyz")

	// Tabulation of neighbor cells, indexed by [direction][parity].
	// Parity is defined as even/0 for latitudes, and odd/1 for longitudes.
	//
	// The indexing is slightly confusing: search [direction][parity] for
	// your source cell, and then base32[index] is your neighbor in that
	// direction (or, you could could think of the direction being inverted).
	neighbors = [][][]byte{
		[][]byte{
			[]byte("p0r21436x8zb9dcf5h7kjnmqesgutwvy"),
			[]byte("bc01fg45238967deuvhjyznpkmstqrwx"),
		},
		[][]byte{
			[]byte("bc01fg45238967deuvhjyznpkmstqrwx"),
			[]byte("p0r21436x8zb9dcf5h7kjnmqesgutwvy"),
		},
		[][]byte{
			[]byte("14365h7k9dcfesgujnmqp0r2twvyx8zb"),
			[]byte("238967debc01fg45kmstqrwxuvhjyznp"),
		},
		[][]byte{
			[]byte("238967debc01fg45kmstqrwxuvhjyznp"),
			[]byte("14365h7k9dcfesgujnmqp0r2twvyx8zb"),
		},
	}

	// This looks like magic, but is a tablulation of which cells are
	// on a discontinuous border, indexed by [direction][parity].
	//
	// For example, the right-most edge of the geohash grid is P,R,X,Z.
	// If you are considering longitudes, and want to go RIGHT, you need to
	// "wrap around" specially.  Hence borders[RIGHT][odd] = "prxz". There
	// are symmetries here: going RIGHT when considering longitudes is the
	// same if you want to go UP along latitudes (borders[UP][even] == "prxz").
	borders = [][][]byte{
		//           even (lat)        odd (lng)
		[][]byte{[]byte("prxz"), []byte("bcfguvyz")}, // UP
		[][]byte{[]byte("bcfguvyz"), []byte("prxz")}, // RIGHT
		[][]byte{[]byte("028b"), []byte("0145hjnp")}, // DOWN
		[][]byte{[]byte("0145hjnp"), []byte("028b")}, // LEFT
	}

	// Note that borders[D][ODD] == borders[D+2][EVEN], and the transformation for neighbors.
)

type LatLng struct {
	Lat float64
	Lng float64
}

func (ll LatLng) String() string {
	return fmt.Sprintf("[%v,%v]", ll.Lat, ll.Lng)
}

// Representing the cell/bounding box that a geohash represents
// (a larger cell for shorter and shorter hashes).
type GeoCell struct {
	TopRight   LatLng
	BottomLeft LatLng
	Center     LatLng
}

func (hash GeoHash) String() string {
	return string(hash)
}

// Returns the precision of the hash (aka the length of the geohash).
func (hash GeoHash) Precision() int {
	return len(hash)
}

func byteToLower(b byte) byte {
	return b | 32
}

// Calculates adjacent geohashes in the given direction.
func (hash GeoHash) Adjacent(dir Direction) GeoHash {
	if len(hash) == 0 {
		return hash
	}

	suffix := byteToLower(hash[len(hash)-1])
	parity := (len(hash) % 2) // 0=even; 1=odd;
	base := hash[:len(hash)-1]
	if bytes.IndexByte(borders[dir][parity], suffix) != -1 {
		base = base.Adjacent(dir)
	}

	adjacent := make(GeoHash, len(base)+1)
	copy(adjacent, base)
	adjacent[len(base)] = base32[bytes.IndexByte(neighbors[dir][parity], suffix)]

	return adjacent
}

// Returns the 8 neighbors for the cell (in clockwise order, starting at the cell above), prepended by the cell itself.
func (hash GeoHash) Neighbors() [9]GeoHash {
	l := hash.Adjacent(Dir_LEFT)
	r := hash.Adjacent(Dir_RIGHT)

	return [9]GeoHash{
		hash,
		hash.Adjacent(Dir_UP),

		r.Adjacent(Dir_UP),
		r,
		r.Adjacent(Dir_DOWN),

		hash.Adjacent(Dir_DOWN),

		l.Adjacent(Dir_DOWN),
		l,
		l.Adjacent(Dir_UP),
	}
}

func refineInterval(interval []float64, bit int) {
	if bit != 0 {
		interval[0] = (interval[0] + interval[1]) / 2
	} else {
		interval[1] = (interval[0] + interval[1]) / 2
	}
}

// Returns the GeoCell coordinates from a geohash.
func (hash GeoHash) ToGeoCell() *GeoCell {
	isEven := true
	lat := []float64{-90, 90}
	lng := []float64{-180, 180}
	for _, b := range hash {
		cd := bytes.IndexByte(base32, byteToLower(b))
		for j := 16; j > 0; j >>= 1 {
			if isEven {
				refineInterval(lng, cd&j)
			} else {
				refineInterval(lat, cd&j)
			}
			isEven = !isEven
		}
	}
	return &GeoCell{
		TopRight:   LatLng{Lat: lat[0], Lng: lng[0]},
		BottomLeft: LatLng{Lat: lat[1], Lng: lng[1]},
		Center: LatLng{
			Lat: (lat[0] + lat[1]) / 2,
			Lng: (lng[0] + lng[1]) / 2,
		},
	}
}

// Create a geohash from a LatLng
func (ll LatLng) ToGeoHash() GeoHash {
	latitude := ll.Lat
	longitude := ll.Lng
	if latitude < -90 || latitude > 90 {
		panic("Invalid latitude")
	}
	if longitude < -180 || longitude > 180 {
		panic("Invalid longitude")
	}

	// TODO(pete) support arbitrary precision passed in, with correct rounding!
	precision := 12

	lat := []float64{-90, 90}
	lng := []float64{-180, 180}
	geohash := make(GeoHash, 0, precision)

	// Modified within the loop as we process base32 characters
	isEven := true
	bit := 16
	ch := 0

	for len(geohash) < precision {
		if isEven {
			mid := (lng[0] + lng[1]) / 2
			if longitude >= mid {
				ch |= bit
				lng[0] = mid
			} else {
				lng[1] = mid
			}
		} else {
			mid := (lat[0] + lat[1]) / 2
			if latitude >= mid {
				ch |= bit
				lat[0] = mid
			} else {
				lat[1] = mid
			}
		}
		isEven = !isEven
		bit >>= 1
		if bit == 0 {
			geohash = append(geohash, base32[ch])
			bit = 16
			ch = 0
		}
	}
	return geohash
}
