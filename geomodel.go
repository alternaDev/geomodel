/*
    Copyright 2012 Alexander Yngling

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/

package geomodel

import "math"

const (
	GEOCELL_GRID_SIZE      = 4
	GEOCELL_ALPHABET       = "0123456789abcdef"
	MAX_GEOCELL_RESOLUTION = 13 // The maximum *practical* geocell resolution.
)

func GeoCell(lat, lon float64, resolution int) string {
	north := 90.0
	south := -90.0
	east := 180.0
	west := -180.0
	cell := make([]byte, resolution, resolution)

	for i := 0; i < resolution; i++ {
		subcellLonSpan := (east - west) / GEOCELL_GRID_SIZE
		subcellLatSpan := (north - south) / GEOCELL_GRID_SIZE

		x := int(math.Min(GEOCELL_GRID_SIZE*(lon-west)/(east-west), GEOCELL_GRID_SIZE-1))
		y := int(math.Min(GEOCELL_GRID_SIZE*(lat-south)/(north-south), GEOCELL_GRID_SIZE-1))

		pos := (y&2)<<2 | (x&2)<<1 | (y&1)<<1 | (x&1)<<0
		cell[i] = GEOCELL_ALPHABET[pos]

		south += subcellLatSpan * float64(y)
		north = south + subcellLatSpan

		west += subcellLonSpan * float64(x)
		east = west + subcellLonSpan
	}

	return string(cell)
}

func GeoCells(lat, lon float64, resolution int) []string {
	g := GeoCell(lat, lon, resolution)
	cells := make([]string, len(g), len(g))
	for i := 0; i < resolution; i++ {
		cells[i] = g[0 : i+1]
	}
	return cells
}
