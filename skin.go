package nude

type Skin struct {
	id      int
	skin    bool
	region  int
	X       int
	Y       int
	chekced bool
}

type SkinMap []*Skin
type SkinMapList []SkinMap

//
// sort interface
//

func (sml SkinMapList) Len() int {
	return len(sml)
}

func (sml SkinMapList) Swap(i, j int) {
	sml[i], sml[j] = sml[j], sml[i]
}

func (sml SkinMapList) Less(i, j int) bool {
	return len(sml[i]) > len(sml[j])
}
