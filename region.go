package nude

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

func (r Region) leftMost() *Pixel {
	minX := 1000000
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
	maxX := -1
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
	minY := 1000000
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
	maxY := -1
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
	left := r.leftMost()
	right := r.rightMost()
	upper := r.upperMost()
	lower := r.lowerMost()
	total := 0
	skin := 0
	var w int
	var h int
	var inclination float64
	// left-upper
	w = upper.X - left.X
	h = left.Y - upper.Y
	inclination = float64(h) / float64(w)
	for y := upper.Y; y < left.Y; y++ {
		xx := float64(y) / inclination
		for x := left.X; x < upper.X; x++ {
			if float64(x) >= xx {
				skin = skin + 1
			}
			total = total + 1
		}
	}
	// upper-right
	w = right.X - upper.X
	h = right.Y - upper.Y
	inclination = float64(h) / float64(w)
	for y := upper.Y; y < right.Y; y++ {
		xx := float64(y) / inclination
		for x := upper.X; x < right.X; x++ {
			if float64(x) <= xx {
				skin = skin + 1
			}
			total = total + 1
		}
	}
	// left-lower
	w = lower.X - left.X
	h = left.Y - lower.Y
	inclination = float64(h) / float64(w)
	for y := left.Y; y < lower.Y; y++ {
		xx := float64(y) / inclination
		for x := left.X; x < lower.X; x++ {
			if float64(x) >= xx {
				skin = skin + 1
			}
			total = total + 1
		}
	}
	// lower-right
	w = right.X - lower.X
	h = right.Y - lower.Y
	inclination = float64(h) / float64(w)
	for y := right.Y; y < lower.Y; y++ {
		xx := float64(y) / inclination
		for x := lower.X; x < right.X; x++ {
			if float64(x) <= xx {
				skin = skin + 1
			}
			total = total + 1
		}
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
