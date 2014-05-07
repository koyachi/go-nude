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

//
// sort interface
//

func (sml Regions) Len() int {
	return len(sml)
}

func (sml Regions) Swap(i, j int) {
	sml[i], sml[j] = sml[j], sml[i]
}

func (sml Regions) Less(i, j int) bool {
	return len(sml[i]) > len(sml[j])
}
