package geomodel

import "testing"
import "sort"

/*type LocationCapable interface {
	Latitude() float64
	Longitude() float64
	Key() string
	Geocells() []string
}*/

type Place struct {
  lat, lon float64
  key string
  geocells []string
}

func (p Place) Latitude() float64 {
  return p.lat
}

func (p Place) Longitude() float64 {
  return p.lon
}

func (p Place) Key() string {
  return p.key
}

func (p Place) Geocells() []string {
  return p.geocells
}

func TestProximityFetch(t *testing.T) {

  var places []LocationCapable = []LocationCapable {Place{54, 8, "1", GeoCells(54, 8, 10)},Place{50, 8, "1", GeoCells(50, 8, 10)},Place{49, 8, "1", GeoCells(49, 8, 10)},Place{48, 8, "1", GeoCells(48, 8, 10)},Place{47, 8, "1", GeoCells(47, 8, 10)}}

  ProximityFetch(50, 8, 20, 300, func(cells []string) []LocationCapable {
    var result []LocationCapable = make([]LocationCapable, 0)

    for _, place := range places {
      var added bool = false
      for _, c := range place.Geocells() {
        index := sort.SearchStrings(cells, c)
        if index < len(cells) {
          if place != nil {
            result = append(result, place)
          }
        }
      }
      if added {
        break
      }
    }

    return result
  }, 10)

  // ProximityFetch(lat, lon float64, maxResults int, maxDistance float64, search RepositorySearch, maxResolution int) []LocationCapable
}
