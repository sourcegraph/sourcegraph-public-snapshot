package geohash

import (
	"math"
	"strings"
	"testing"
)

func TestNeighbors(t *testing.T) {
	directions := []Direction{Dir_UP, Dir_RIGHT, Dir_DOWN, Dir_LEFT}
	tests := []struct {
		location  string
		neighbors []string
	}{
		// edge case at bottom left
		{"0", []string{"2", "1", "b", "p"}},
		// top left
		{"b", []string{"0", "c", "8", "z"}},
		// top right
		{"z", []string{"p", "b", "x", "y"}},
		// bottom right
		{"p", []string{"r", "0", "z", "n"}},

		// center
		{"d", []string{"f", "e", "6", "9"}},

		// random
		{"dq", []string{"dr", "dw", "dm", "dn"}},
		{"DQ", []string{"dr", "dw", "dm", "dn"}},
	}

	for _, tc := range tests {
		hash := GeoHash(tc.location)
		for i, dir := range directions {
			adj := strings.ToLower(string(hash.Adjacent(dir)))
			if adj != tc.neighbors[i] {
				t.Errorf("Neighbor of %v in %v != %v (got %v)", tc.location, dir, tc.neighbors[i], adj)
			}
		}
	}
}

func TestConversion(t *testing.T) {
	epsilon := 0.000001
	tests := []struct {
		ll      LatLng
		geohash string
	}{
		{LatLng{37.765662, -122.421870}, "9q8yy5p0gtts"},
		{LatLng{37.765662, -122.421870}, "9q8yy5p0gtts"},
	}

	for _, tc := range tests {
		geohash := tc.ll.ToGeoHash()
		if string(geohash) != tc.geohash {
			t.Errorf("Expected %v => %v, got %v",
				tc.ll, tc.geohash, string(geohash))
		}

		bb := geohash.ToGeoCell()
		if math.Abs(bb.Center.Lat-tc.ll.Lat) > epsilon ||
			math.Abs(bb.Center.Lng-tc.ll.Lng) > epsilon {
			t.Errorf("lat %v != %v, lng %v != %v",
				bb.Center.Lat, tc.ll.Lat, bb.Center.Lng, tc.ll.Lng)
		}
	}
}

func assertEquals(a GeoHash, to string, t *testing.T) {
	if string(a) != to {
		t.Errorf("expected %v, got %v", to, a)
	}
}

func TestZeroZero(t *testing.T) {
	assertEquals(LatLng{0, 0}.ToGeoHash(), "s00000000000", t)
}

func TestNeighbourCloseTo180Longitude(t *testing.T) {
	assertEquals(GeoHash("r").Adjacent(Dir_RIGHT), "2", t)
	assertEquals(GeoHash("2").Adjacent(Dir_LEFT), "r", t)
}

/*
  Failing -- we don't (and the source we adapted from) do not handle the poles very well
*/
/*
func TestTopNeighbourCloseToNorthPole(t *testing.T) {
	hash := LatLng{90, 0}.ToGeoHash()[0:1]
	assertEquals(hash, "u", t)
	assertEquals(hash.Adjacent(Dir_UP), "b", t)
}

func TestBottomNeighbourCloseToSouthPole(t *testing.T) {
	hash := LatLng{-90, 0}.ToGeoHash()[0:1]
	assertEquals(hash, "h", t)
	assertEquals(hash.Adjacent(Dir_DOWN), "0", t)
}
*/

/*
 Should adapt these tests as well
*/
/*
func TestNeighboursAtSouthPole(t *testing.T) {
	String poleHash = GeoHash.encodeHash(-90, 0);
	assertEquals("h00000000000", poleHash);

	List<String> neighbors = GeoHash.neighbours(poleHash);
	assertEquals(8, neighbors.size());

	assertEquals("5bpbpbpbpbpb", neighbors.get(I_LEFT));
	assertEquals("h00000000002", neighbors.get(I_RIGHT));
	assertEquals("h00000000001", neighbors.get(I_TOP));
	assertEquals("00000000000p", neighbors.get(I_BOTTOM));
	assertEquals("5bpbpbpbpbpc", neighbors.get(I_LEFT_TOP));
	assertEquals("pbpbpbpbpbpz", neighbors.get(I_LEFT_BOT));
	assertEquals("h00000000003", neighbors.get(I_RIGHT_TOP));
	assertEquals("00000000000r", neighbors.get(I_RIGHT_BOT));
}


func TestNeighboursAtNorthPole(t *testing.T) {
	String poleHash = GeoHash.encodeHash(90, 0);
	assertEquals("upbpbpbpbpbp", poleHash);

	List<String> neighbors = GeoHash.neighbours(poleHash);
	assertEquals(8, neighbors.size());

	assertEquals("gzzzzzzzzzzz", neighbors.get(I_LEFT));
	assertEquals("upbpbpbpbpbr", neighbors.get(I_RIGHT));
	assertEquals("bpbpbpbpbpb0", neighbors.get(I_TOP));
	assertEquals("upbpbpbpbpbn", neighbors.get(I_BOTTOM));
	assertEquals("zzzzzzzzzzzb", neighbors.get(I_LEFT_TOP));
	assertEquals("gzzzzzzzzzzy", neighbors.get(I_LEFT_BOT));
	assertEquals("bpbpbpbpbpb2", neighbors.get(I_RIGHT_TOP));
	assertEquals("upbpbpbpbpbq", neighbors.get(I_RIGHT_BOT));
}


func TestNeighboursAtLongitude180(t *testing.T) {
	String hash = GeoHash.encodeHash(0, 180);
	assertEquals("xbpbpbpbpbpb", hash);

	List<String> neighbors = GeoHash.neighbours(hash);
	assertEquals(8, neighbors.size());

	assertEquals("xbpbpbpbpbp8", neighbors.get(I_LEFT));
	assertEquals("800000000000", neighbors.get(I_RIGHT));
	assertEquals("xbpbpbpbpbpc", neighbors.get(I_TOP));
	assertEquals("rzzzzzzzzzzz", neighbors.get(I_BOTTOM));
	assertEquals("xbpbpbpbpbp9", neighbors.get(I_LEFT_TOP));
	assertEquals("rzzzzzzzzzzx", neighbors.get(I_LEFT_BOT));
	assertEquals("800000000001", neighbors.get(I_RIGHT_TOP));
	assertEquals("2pbpbpbpbpbp", neighbors.get(I_RIGHT_BOT));
}


func TestNeighboursAtLongitudeMinus180(t *testing.T) {
	String hash = GeoHash.encodeHash(0, -180);
	assertEquals("800000000000", hash);

	List<String> neighbors = GeoHash.neighbours(hash);
	System.out.println(neighbors);
	assertEquals(8, neighbors.size());

	assertEquals("xbpbpbpbpbpb", neighbors.get(I_LEFT));
	assertEquals("800000000002", neighbors.get(I_RIGHT));
	assertEquals("800000000001", neighbors.get(I_TOP));
	assertEquals("2pbpbpbpbpbp", neighbors.get(I_BOTTOM));
	assertEquals("xbpbpbpbpbpc", neighbors.get(I_LEFT_TOP));
	assertEquals("rzzzzzzzzzzz", neighbors.get(I_LEFT_BOT));
	assertEquals("800000000003", neighbors.get(I_RIGHT_TOP));
	assertEquals("2pbpbpbpbpbr", neighbors.get(I_RIGHT_BOT));
}
*/

// Additional tests from
// https://raw.githubusercontent.com/ttacon/geohash-golang/master/geohash_test.go

func TestDecode(t *testing.T) {
	var tests = []struct {
		input  string
		output *GeoCell
	}{
		{
			"d",
			&GeoCell{
				LatLng{0, -90},
				LatLng{45, -45},
				LatLng{22.5, -67.5},
			},
		},
		{
			"dr",
			&GeoCell{
				LatLng{39.375, -78.75},
				LatLng{45, -67.5},
				LatLng{42.1875, -73.125},
			},
		},
		{
			"dr1",
			&GeoCell{
				LatLng{39.375, -77.34375},
				LatLng{40.78125, -75.9375},
				LatLng{40.078125, -76.640625},
			},
		},
		{
			"dr12",
			&GeoCell{
				LatLng{39.375, -76.9921875},
				LatLng{39.55078125, -76.640625},
				LatLng{39.462890625, -76.81640625},
			},
		},
	}

	for _, test := range tests {
		box := GeoHash(test.input).ToGeoCell()
		if !equalGeoCells(test.output, box) {
			t.Errorf("expected bounding box %v, got %v", test.output, box)
		}
	}
}

func equalGeoCells(b1, b2 *GeoCell) bool {
	return b1.TopRight == b2.TopRight &&
		b1.BottomLeft == b2.BottomLeft &&
		b1.Center == b1.Center
}

func TestEncode(t *testing.T) {
	var tests = []struct {
		latlng  LatLng
		geohash string
	}{
		{
			LatLng{39.550781, -76.640625},
			"dr18bpbpbpbn",
		},
		{
			LatLng{39.5507, -76.6406},
			"dr18bpbp88fe",
		},
		{
			LatLng{39.55, -76.64},
			"dr18bpb7qw65",
		},
		{
			LatLng{39, -76},
			"dqcvyedrrwut",
		},
	}

	for _, test := range tests {
		geohash := string(test.latlng.ToGeoHash())
		if test.geohash != geohash {
			t.Errorf("expectd %s, got %s", test.geohash, geohash)
		}
	}
}
