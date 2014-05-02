package nude

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"math"
	"os"
	"path/filepath"
	"sort"
)

func IsNude(imageFilePath string) (result bool, err error) {
	path, err := filepath.Abs(imageFilePath)
	if err != nil {
		return false, err
	}
	n := New(path)
	result, err = n.Parse()

	return
}

type Skin struct {
	id      int
	skin    bool
	region  int
	x       int
	y       int
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

type Nude struct {
	filePath        string
	image           image.Image
	width           int
	height          int
	totalPixels     int
	skinMap         SkinMap
	skinRegions     SkinMapList
	detectedRegions SkinMapList
	mergeRegions    [][]int
	lastFrom        int
	lastTo          int
	message         string
	result          bool
}

func New(path string) *Nude {
	nude := &Nude{
		filePath: path,
	}
	return nude
}

func (nude *Nude) Parse() (result bool, err error) {
	reader, err := os.Open(nude.filePath)
	if err != nil {
		return false, err
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return false, err
	}
	bounds := img.Bounds()
	nude.image = img
	nude.width = bounds.Size().X
	nude.height = bounds.Size().Y
	nude.lastFrom = -1
	nude.lastTo = -1
	nude.totalPixels = nude.width * nude.height

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		width := bounds.Size().X
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := nude.image.At(x, y).RGBA()
			normR := r / 256
			normG := g / 256
			normB := b / 256
			index := x + y*width + 1

			if !classifySkin(normR, normG, normB) {
				nude.skinMap = append(nude.skinMap, &Skin{index, false, 0, x, y, false})
			} else {
				nude.skinMap = append(nude.skinMap, &Skin{index, true, 0, x, y, false})

				region := -1
				checkIndexes := []int{
					index - 2,
					index - width - 2,
					index - width - 1,
					index - width,
				}
				checker := false

				for _, checkIndex := range checkIndexes {
					if checkIndex < 0 {
						continue
					}
					skin := nude.skinMap[checkIndex]
					if skin != nil && skin.skin {
						if skin.region != region &&
							region != -1 &&
							nude.lastFrom != region &&
							nude.lastTo != skin.region {
							nude.addMerge(region, skin.region)
						}
						region = nude.skinMap[checkIndex].region
						checker = true
					}
				}

				if !checker {
					nude.skinMap[index-1].region = len(nude.detectedRegions)
					nude.detectedRegions = append(nude.detectedRegions, []*Skin{nude.skinMap[index-1]})
					continue
				} else {
					if region > -1 {
						if len(nude.detectedRegions) >= region {
							nude.detectedRegions = append(nude.detectedRegions, SkinMap{})
						}
						nude.skinMap[index-1].region = region
						nude.detectedRegions[region] = append(nude.detectedRegions[region], nude.skinMap[index-1])
					}
				}
			}
		}
	}

	nude.merge(nude.detectedRegions, nude.mergeRegions)
	nude.analyzeRegions()

	return nude.result, err
}

func (nude *Nude) addMerge(from, to int) {
	nude.lastFrom = from
	nude.lastTo = to

	fromIndex := -1
	toIndex := -1

	for index, region := range nude.mergeRegions {
		for _, regionIndex := range region {
			if regionIndex == from {
				fromIndex = index
			}
			if regionIndex == to {
				toIndex = index
			}
		}
	}

	if fromIndex != -1 && toIndex != -1 {
		if fromIndex != toIndex {
			fromRegion := nude.mergeRegions[fromIndex]
			toRegion := nude.mergeRegions[toIndex]
			region := append(fromRegion, toRegion...)
			nude.mergeRegions[fromIndex] = region
			nude.mergeRegions = append(nude.mergeRegions[:toIndex], nude.mergeRegions[toIndex+1:]...)
		}
		return
	}

	if fromIndex == -1 && toIndex == -1 {
		nude.mergeRegions = append(nude.mergeRegions, []int{from, to})
		return
	}

	if fromIndex != -1 && toIndex == -1 {
		nude.mergeRegions[fromIndex] = append(nude.mergeRegions[fromIndex], to)
		return
	}

	if fromIndex == -1 && toIndex != -1 {
		nude.mergeRegions[toIndex] = append(nude.mergeRegions[toIndex], from)
		return
	}

}

// function for merging detected regions
func (nude *Nude) merge(detectedRegions SkinMapList, mergeRegions [][]int) {
	var newDetectedRegions SkinMapList

	// merging detected regions
	for index, region := range mergeRegions {
		if len(newDetectedRegions) >= index {
			newDetectedRegions = append(newDetectedRegions, SkinMap{})
		}
		for _, r := range region {
			newDetectedRegions[index] = append(newDetectedRegions[index], detectedRegions[r]...)
			detectedRegions[r] = SkinMap{}
		}
	}

	// push the rest of the regions to the newDetectedRegions array
	// (regions without merging)
	for _, region := range detectedRegions {
		if len(region) > 0 {
			newDetectedRegions = append(newDetectedRegions, region)
		}
	}

	// clean up
	nude.clearRegions(newDetectedRegions)
}

// clean up function
// only push regions which are bigger than a specific amount to the final resul
func (nude *Nude) clearRegions(detectedRegions SkinMapList) {
	for _, region := range detectedRegions {
		if len(region) > 30 {
			nude.skinRegions = append(nude.skinRegions, region)
		}
	}
}

func (nude *Nude) analyzeRegions() bool {
	// if there are less than 3 regions
	if len(nude.skinRegions) < 3 {
		nude.message = fmt.Sprintf("Less than 3 skin regions (%v)", len(nude.skinRegions))
		nude.result = false
		return nude.result
	}

	// sort the skinRegions
	sort.Sort(nude.skinRegions)
	//sort.Reverse(nude.skinRegions)

	// count total skin pixels
	var totalSkin float64
	for _, region := range nude.skinRegions {
		totalSkin += float64(len(region))
	}

	// check if there are more than 15% skin pixel in the image
	totalSkinParcentage := totalSkin / float64(nude.totalPixels) * 100
	if totalSkinParcentage < 15 {
		// if the parcentage lower than 15, it's not nude!
		nude.message = fmt.Sprintf("Total skin parcentage lower than 15 (%v%%)", totalSkinParcentage)
		nude.result = false
		return nude.result
	}

	// check if the largest skin region is less than 35% of the total skin count
	// AND if the second largest region is less than 30% of the total skin count
	// AND if the third largest region is less than 30% of the total skin count
	biggestRegionParcentage := float64(len(nude.skinRegions[0])) / totalSkin * 100
	secondLargeRegionParcentage := float64(len(nude.skinRegions[1])) / totalSkin * 100
	thirdLargesRegionParcentage := float64(len(nude.skinRegions[2])) / totalSkin * 100
	if biggestRegionParcentage < 35 &&
		secondLargeRegionParcentage < 30 &&
		thirdLargesRegionParcentage < 30 {
		nude.message = "Less than 35%, 30%, 30% skin in the biggest regions"
		nude.result = false
		return nude.result
	}

	// check if the number of skin pixels in the largest region is less than 45% of the total skin count
	if biggestRegionParcentage < 45 {
		nude.message = fmt.Sprintf("The biggest region contains less than 45%% (%v)", biggestRegionParcentage)
		nude.result = false
		return nude.result
	}

	// TODO:
	// build the bounding polygon by the regions edge values:
	// Identify the leftmost, the uppermost, the rightmost, and the lowermost skin pixels of the three largest skin regions.
	// Use these points as the corner points of a bounding polygon.

	// TODO:
	// check if the total skin count is less than 30% of the total number of pixels
	// AND the number of skin pixels within the bounding polygon is less than 55% of the size of the polygon
	// if this condition is true, it's not nude.

	// TODO: include bounding polygon functionality
	// if there are more than 60 skin regions and the average intensity within the polygon is less than 0.25
	// the image is not nude
	if len(nude.skinRegions) > 60 {
		nude.message = fmt.Sprintf("More than 60 skin regions (%v)", len(nude.skinRegions))
		nude.result = false
		return nude.result
	}

	// otherwise it is nude
	nude.result = true
	return nude.result
}

// A Survey on Pixel-Based Skin Color Detection Techniques
func classifySkin(r, g, b uint32) bool {
	rgbClassifier := r > 95 &&
		g > 40 && g < 100 &&
		b > 20 &&
		maxRgb(r, g, b)-minRgb(r, g, b) > 15 &&
		math.Abs(float64(r-g)) > 15 &&
		r > g &&
		r > b

	nr, ng, _ := toNormalizedRgb(r, g, b)
	normalizedRgbClassifier := nr/ng > 1.185 &&
		(float64(r*b))/math.Pow(float64(r+g+b), 2) > 0.107 &&
		(float64(r*g))/math.Pow(float64(r+g+b), 2) > 0.112

	h, s, _ := toHsv(r, g, b)
	hsvClassifier := h > 0 &&
		h < 35 &&
		s > 0.23 &&
		s < 0.68

	// ycc doesnt work

	result := rgbClassifier || normalizedRgbClassifier || hsvClassifier
	return result
}

func maxRgb(r, g, b uint32) float64 {
	return math.Max(math.Max(float64(r), float64(g)), float64(b))
}

func minRgb(r, g, b uint32) float64 {
	return math.Min(math.Min(float64(r), float64(g)), float64(b))
}

func toNormalizedRgb(r, g, b uint32) (nr, ng, nb float64) {
	sum := float64(r + g + b)
	nr = float64(r) / sum
	ng = float64(g) / sum
	nb = float64(b) / sum

	return nr, ng, nb
}

func toHsv(r, g, b uint32) (h, s, v float64) {
	h = 0.0
	sum := float64(r + g + b)
	max := maxRgb(r, g, b)
	min := minRgb(r, g, b)
	diff := max - min

	if max == float64(r) {
		h = float64(g-b) / diff
	} else if max == float64(g) {
		h = 2 + float64(g-r)/diff
	} else {
		h = 4 + float64(r-g)/diff
	}

	h *= 60
	if h < 0 {
		h += 360
	}

	s = 1.0 - 3.0*(min/sum)
	v = (1.0 / 3.0) * max

	return h, s, v
}
