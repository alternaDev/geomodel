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
import "sort"
import "log"

const (
	GEOCELL_GRID_SIZE      = 4
	GEOCELL_ALPHABET       = "0123456789abcdef"
	MAX_GEOCELL_RESOLUTION = 13 // The maximum *practical* geocell resolution.
)

var (
	NORTHWEST = []int{-1,1}
	NORTH 	  = []int{0,1}
	NORTHEAST = []int{1,1}
	EAST      = []int{1,0}
	SOUTHEAST = []int{1,-1}
	SOUTH     = []int{0,-1}
	SOUTHWEST = []int{-1,-1}
	WEST      = []int{-1,0}
)

type LocationCapable interface {
	Latitude() float64
	Longitude() float64
	Key() string
	Geocells() []string
}

type BoundingBox struct {
	latNE float64
	lonNE float64
	latSW float64
	lonSW float64
}

func NewBoundingBox(north, east, south, west float64) BoundingBox {
	var north_, south_ float64
	if south > north {
		south_ = north
		north_ = south
	} else {
		south_ = south
		north_ = north
	}

	return BoundingBox{north_, east, south_, west}
}

type RepositorySearch func([]string) []LocationCapable

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

func DegToRad(val float64) float64 {
	return (math.Pi / 180) * val
}

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	var p1lat = DegToRad(lat1)
	var p1lon = DegToRad(lon1)
	var p2lat = DegToRad(lat2)
	var p2lon = DegToRad(lon2)
	return 6378135 * math.Acos(math.Sin(p1lat) * math.Sin(p2lat) + math.Cos(p1lat) * math.Cos(p2lat) * math.Cos(p2lon - p1lon))
}

func adjacent(cell string, dir []int) string {
	var dx int = dir[0]
	var dy int = dir[1]
	var i  int = len(cell) - 1

	for i >= 1 && (dx != 0 || dy != 0) {
		var l []int = subdivXY(cell[i])
		var x int = l[0]
		var y int = l[1]

		// Horizontal
		if dx == -1 {
			if x == 0 {
				x = GEOCELL_GRID_SIZE - 1
			} else {
				x--
				dx = 0
			}
		} else if dx == 1 {
			if x == GEOCELL_GRID_SIZE - 1 {
				x = 0
			} else {
				x++
				dx = 0
			}
		}

		// Vertical
		if dy == 1 {
			if y == GEOCELL_GRID_SIZE - 1 {
				y = 0
			} else {
				y++
				dy = 0
			}
		} else if dy == -1 {
			if y == 0 {
				y = GEOCELL_GRID_SIZE - 1
			} else {
				y--
				dy = 0
			}
		}

		var l2 []int = []int{x, y}
		cell = string(append([]byte(cell[:i - 1]), subdivChar(l2)))
		cell = string(append([]byte(cell), []byte(cell[i + 1:])...))
		i--
	}

	if dy != 0 {
		return ""
	}

	return cell
}

func distanceSortedEdges(cells []string, lat, lon float64) []IntArrayDoubleTuple {
	var boxes []BoundingBox = make([]BoundingBox, len(cells))
	for _, cell := range cells {
		boxes = append(boxes, computeBox(cell))
	}

	var maxNorth float64 = -math.MaxFloat64
	var maxEast  float64 = -math.MaxFloat64
	var maxSouth float64 = -math.MaxFloat64
	var maxWest  float64 = -math.MaxFloat64

	for _, box := range boxes {
		maxNorth = math.Max(maxNorth, box.latNE)
		maxEast  = math.Max(maxEast, box.lonNE)
		maxSouth = math.Max(maxSouth, box.latSW)
		maxWest  = math.Max(maxWest, box.lonSW)
	}

	result := make([]IntArrayDoubleTuple, 4)
	result[0] = IntArrayDoubleTuple{SOUTH, Distance(maxSouth, lon,     lat, lon)}
	result[1] = IntArrayDoubleTuple{NORTH, Distance(maxNorth, lon,     lat, lon)}
	result[2] = IntArrayDoubleTuple{WEST,  Distance(lat,      maxWest, lat, lon)}
	result[3] = IntArrayDoubleTuple{EAST,  Distance(maxSouth, maxEast, lat, lon)}

	sort.Sort(ByDistanceIA(result))

	return result
}

func subdivXY(char_ uint8) []int {
	var charI int = int(GEOCELL_ALPHABET[char_])
	return []int{(charI & 4) >> 1 | (charI & 1) >> 0, (charI & 8) >> 2 | (charI & 2) >> 1}
}

func subdivChar(pos []int) uint8 {
	return GEOCELL_ALPHABET[(pos[1] & 2) << 2 | (pos[0] & 2) << 1 | (pos[1] & 1) << 1 | (pos[0] & 1) << 0]
}

func computeBox(cell string) BoundingBox {
	var bbox BoundingBox
	if cell == "" {
		return bbox
	}

	bbox = NewBoundingBox(90.0, 180.0, -90.0, -180.0)
	for len(cell) > 0 {
		var subcellLonSpan float64 = (bbox.lonNE - bbox.lonSW) / GEOCELL_GRID_SIZE
		var subcellLatSpan float64 = (bbox.latNE - bbox.latSW) / GEOCELL_GRID_SIZE

		var l []int = subdivXY(cell[0])
		var x int = l[0]
		var y int = l[1]

		bbox = NewBoundingBox(bbox.latSW + subcellLatSpan * (float64(y) + 1),
													bbox.lonSW + subcellLonSpan * (float64(x) + 1),
												  bbox.latSW + subcellLatSpan * float64(y),
												  bbox.lonSW + subcellLonSpan * float64(x))
		cell = cell[1:]
	}

	return bbox
}

type IntArrayDoubleTuple struct {
	first []int
	second float64
}

type ByDistanceIA []IntArrayDoubleTuple
func (a ByDistanceIA) Len() int 					{ return len(a) }
func (a ByDistanceIA) Swap(i, j int)		  { a[i], a[j] = a[j], a[i] }
func (a ByDistanceIA) Less(i, j int) bool { return a[i].second < a[j].second }

func deleteRecords(data []string, remove []string) []string {
    w := 0 // write index

loop:
    for _, x := range data {
        for _, id := range remove {
            if id == x {
                continue loop
            }
        }
        data[w] = x
        w++
    }
    return data[:w]
}

func contains(data []LocationComparableTuple, e LocationComparableTuple) bool {
	for _, a := range data {
		if a.first == e.first && a.second == e.second {
			return true
		}
	}
	return false
}

type LocationComparableTuple struct {
	first LocationCapable
	second float64
}

type ByDistance []LocationComparableTuple
func (a ByDistance) Len() int 					{ return len(a) }
func (a ByDistance) Swap(i, j int)		  { a[i], a[j] = a[j], a[i] }
func (a ByDistance) Less(i, j int) bool { return a[i].second < a[j].second }

func ProximityFetch(lat, lon float64, maxResults int, maxDistance float64, search RepositorySearch, maxResolution int) []LocationCapable {
	var results []LocationComparableTuple

	var curContainingGeocell string = GeoCell(lat, lon, maxResolution)

	var searchedCells []string = make([]string, maxResults, maxResults * 2)

	var curGeocells []string = make([]string, maxResults, maxResults * 2)
	curGeocells = append(curGeocells, curContainingGeocell)
	var closestPossibleNextResultDist float64 = 0


	var noDirection  = []int{0, 0}

	var sortedEdgeDistances []IntArrayDoubleTuple
	sortedEdgeDistances = append(sortedEdgeDistances, IntArrayDoubleTuple{noDirection, 0})

	for len(curGeocells) != 0 {
		closestPossibleNextResultDist = sortedEdgeDistances[0].second
		if(maxDistance > 0 && closestPossibleNextResultDist > maxDistance) {
			break
		}

		var curTempUnique = deleteRecords(curGeocells, searchedCells)


		var curGeocellsUnique = curTempUnique

		var newResultEntities = search(curGeocellsUnique)

		searchedCells = append(searchedCells, curGeocells...)


		// Begin storing distance from the search result entity to the
    // search center along with the search result itself, in a tuple.
		var newResults []LocationComparableTuple = make([]LocationComparableTuple, len(newResultEntities), len(newResultEntities) * 2)
		for _, entity := range newResultEntities {
			newResults = append(newResults, LocationComparableTuple{entity, Distance(lat, lon, entity.Latitude(), entity.Longitude())})
		}

		sort.Sort(ByDistance(newResults))
		newResults = newResults[0:int(math.Min(float64(maxResults), float64(len(newResults))))]

		// Merge new_results into results
		for _, tuple := range newResults {
			// contains method will check if entity in tuple have same key
			if(!contains(results, tuple)) {
				results = append(results, tuple)
			}
		}

		sort.Sort(ByDistance(results))
		results = results[0:int(math.Min(float64(maxResults), float64(len(results))))]

		sortedEdgeDistances = distanceSortedEdges(curGeocells, lat, lon)

		if len(results) == 0 || len(curGeocells) == 4 {
			/* Either no results (in which case we optimize by not looking at
         adjacents, go straight to the parent) or we've searched 4 adjacent
         geocells, in which case we should now search the parents of those
         geocells.*/
			curContainingGeocell = curContainingGeocell[:int(math.Max(float64(len(curContainingGeocell)) - 1, float64(0)))]

			if len(curContainingGeocell) == 0 {
				break
			}

			var oldCurGeocells []string = curGeocells
			curGeocells = make([]string, 4)

			for _, cell := range oldCurGeocells {
				if len(cell) > 0 {
					var newCell string = cell[:len(cell) - 1]
					i := sort.SearchStrings(curGeocells, newCell)
					if !(i < len(curGeocells) && curGeocells[i] == newCell) {
						curGeocells = append(curGeocells, newCell)
					}
				}
			}

			if len(curGeocells) == 0 {
				break
			}
		} else if len(curGeocells) == 1 {
			var nearestEdge []int = sortedEdgeDistances[0].first
			curGeocells = append(curGeocells, adjacent(curGeocells[0], nearestEdge))
		} else if len(curGeocells) == 2 {
			var nearestEdge []int = distanceSortedEdges([]string{curContainingGeocell}, lat, lon)[0].first
			var perpendicularNearestEdge []int = []int{0, 0}

			if(nearestEdge[0] == 0) {
				for _, edgeDistance := range sortedEdgeDistances {
					if edgeDistance.first[0] != 0 {
						perpendicularNearestEdge = edgeDistance.first
						break
					}
				}
			} else {
				for _, edgeDistance := range sortedEdgeDistances {
					if edgeDistance.first[0] == 0 {
						perpendicularNearestEdge = edgeDistance.first
						break
					}
				}
			}

			var tempCells []string = make([]string, len(curGeocells))

			for _, cell := range curGeocells {
				tempCells = append(tempCells, adjacent(cell, perpendicularNearestEdge))
			}

			curGeocells = append(curGeocells, tempCells...)
		}

		if len(results) < maxResults {
			// Keep Searchin!
			log.Print("%d results found but want %d results, continuing...", len(results), maxResults)
			continue
		}

		// Found things!
		log.Print("%d results found.", len(results))

		var currentFarthestReturnableResultDist float64 = Distance(lat, lon, results[maxResults - 1].first.Latitude(), results[maxResults - 1].first.Longitude())

		if(closestPossibleNextResultDist >= currentFarthestReturnableResultDist) {
			// Done
			log.Print("DONE next result at least %d away, current farthest is %d dist.", closestPossibleNextResultDist, currentFarthestReturnableResultDist)
			break
		}

		log.Print("next result at least %d away, current farthest is %d dist", closestPossibleNextResultDist, currentFarthestReturnableResultDist)

	}

	var result []LocationCapable = make([]LocationCapable, maxResults)

	for _, entry := range results[0:int(math.Min(float64(maxResults), float64(len(results))))] {
		if(maxDistance == 0 || entry.second < maxDistance) {
			result = append(result, entry.first)
		}
	}

	return result
}
