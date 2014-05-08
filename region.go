package nude

type Pixel struct {
	id      int
	isSkin  bool
	region  int
	X       int
	Y       int
	chekced bool
}

type Pixels []*Pixel

type Regions []Pixels

func (r Regions) totalPixels() int {
	var totalSkin int
	for _, pixels := range r {
		totalSkin += len(pixels)
	}
	return totalSkin
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
