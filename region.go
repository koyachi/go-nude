package nude

import (
	"math"
)

type Pixel struct {
	id      int
	isSkin  bool
	region  int
	X       int
	Y       int
	chekced bool
	V       float64 // intesitiy(Value) of HSV
}

// TODO: cache caluculated leftMost, rightMost, upperMost, lowerMost.
type Region []*Pixel

// TODO: optimize
//func (r Region) isSkin(x, y int) bool {
//	for _, pixel := range r {
//		if pixel.isSkin && pixel.X == x && pixel.Y == y {
//			return true
//		}
//	}
//	return false
//}

func (r Region) leftMost() *Pixel {
	minX := math.MaxInt32
	index := 0
	for i, pixel := range r {
		if pixel.X < minX {
			minX = pixel.X
			index = i
		}
	}
	return r[index]
}

func (r Region) rightMost() *Pixel {
	maxX := math.MinInt32
	index := 0
	for i, pixel := range r {
		if pixel.X > maxX {
			maxX = pixel.X
			index = i
		}
	}
	return r[index]
}

func (r Region) upperMost() *Pixel {
	minY := math.MaxInt32
	index := 0
	for i, pixel := range r {
		if pixel.Y < minY {
			minY = pixel.Y
			index = i
		}
	}
	return r[index]
}

func (r Region) lowerMost() *Pixel {
	maxY := math.MinInt32
	index := 0
	for i, pixel := range r {
		if pixel.Y > maxY {
			maxY = pixel.Y
			index = i
		}
	}
	return r[index]
}

func (r Region) skinRateInBoundingPolygon() float64 {
	// build the bounding polygon by the regions edge values:
	// Identify the leftmost, the uppermost, the rightmost, and the lowermost skin pixels of the three largest skin regions.
	// Use these points as the corner points of a bounding polygon.
	left := r.leftMost()
	right := r.rightMost()
	upper := r.upperMost()
	lower := r.lowerMost()
	vertices := []*Pixel{left, upper, right, lower, left}
	total := 0
	skin := 0

	// via http://katsura-kotonoha.sakura.ne.jp/prog/c/tip0002f.shtml
	for _, p1 := range r {
		inPolygon := true
		for i := 0; i < len(vertices)-1; i++ {
			p2 := vertices[i]
			p3 := vertices[i+1]
			n := p1.X*(p2.Y-p3.Y) + p2.X*(p3.Y-p1.Y) + p3.X*(p1.Y-p2.Y)
			if n < 0 {
				inPolygon = false
				break
			}
		}
		if inPolygon && p1.isSkin {
			skin = skin + 1
		}
		total = total + 1
	}
	return float64(skin) / float64(total)
}

func (r Region) averageIntensity() float64 {
	var totalIntensity float64
	for _, pixel := range r {
		totalIntensity = totalIntensity + pixel.V
	}
	return totalIntensity / float64(len(r))
}

type Regions []Region

func (r Regions) totalPixels() int {
	var totalSkin int
	for _, pixels := range r {
		totalSkin += len(pixels)
	}
	return totalSkin
}

func (r Regions) averageIntensity() float64 {
	var totalIntensity float64
	for _, region := range r {
		totalIntensity = totalIntensity + region.averageIntensity()
	}
	return totalIntensity / float64(len(r))
}

//
// sort interface
//

func (r Regions) Len() int {
	return len(r)
}

func (r Regions) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Regions) Less(i, j int) bool {
	return len(r[i]) < len(r[j])
}
