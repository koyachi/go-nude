package nude

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func IsNude(imageFilePath string) (result bool, err error) {
	path, err := filepath.Abs(imageFilePath)
	if err != nil {
		return false, err
	}
	a := New(path)
	result, err = a.Parse()

	return
}

// experimental
func DecodeImage(filePath string) (img image.Image, err error) {
	return decodeImage(filePath)
}

func decodeImage(filePath string) (img image.Image, err error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	last3Strings := strings.ToLower(filePath[len(filePath)-3:])
	last4Strings := strings.ToLower(filePath[len(filePath)-4:])
	if last3Strings == "jpg" || last4Strings == "jpeg" {
		img, err = jpeg.Decode(reader)
	} else if last3Strings == "gif" {
		img, err = gif.Decode(reader)
	} else if last3Strings == "png" {
		img, err = png.Decode(reader)
	} else {
		img = nil
		err = errors.New("unknown format")
	}
	return
}

type Analyzer struct {
	filePath        string
	image           image.Image
	width           int
	height          int
	totalPixels     int
	skinMap         SkinMap
	SkinRegions     SkinMapList
	detectedRegions SkinMapList
	mergeRegions    [][]int
	lastFrom        int
	lastTo          int
	message         string
	result          bool
}

func New(path string) *Analyzer {
	analyzer := &Analyzer{
		filePath: path,
	}
	return analyzer
}

func (a *Analyzer) Parse() (result bool, err error) {
	img, err := decodeImage(a.filePath)
	if err != nil {
		return false, err
	}
	bounds := img.Bounds()
	a.image = img
	a.width = bounds.Size().X
	a.height = bounds.Size().Y
	a.lastFrom = -1
	a.lastTo = -1
	a.totalPixels = a.width * a.height

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		width := bounds.Size().X
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := a.image.At(x, y).RGBA()
			normR := r / 256
			normG := g / 256
			normB := b / 256
			currentIndex := x + y*width
			nextIndex := currentIndex + 1

			if !classifySkin(normR, normG, normB) {
				a.skinMap = append(a.skinMap, &Skin{currentIndex, false, 0, x, y, false})
			} else {
				a.skinMap = append(a.skinMap, &Skin{currentIndex, true, 0, x, y, false})

				region := -1
				checkIndexes := []int{
					nextIndex - 2,
					nextIndex - width - 2,
					nextIndex - width - 1,
					nextIndex - width,
				}
				checker := false

				for _, checkIndex := range checkIndexes {
					if checkIndex < 0 {
						continue
					}
					skin := a.skinMap[checkIndex]
					if skin != nil && skin.skin {
						if skin.region != region &&
							region != -1 &&
							a.lastFrom != region &&
							a.lastTo != skin.region {
							a.addMerge(region, skin.region)
						}
						region = a.skinMap[checkIndex].region
						checker = true
					}
				}

				if !checker {
					a.skinMap[currentIndex].region = len(a.detectedRegions)
					a.detectedRegions = append(a.detectedRegions, []*Skin{a.skinMap[currentIndex]})
					continue
				} else {
					if region > -1 {
						if len(a.detectedRegions) >= region {
							a.detectedRegions = append(a.detectedRegions, SkinMap{})
						}
						a.skinMap[currentIndex].region = region
						a.detectedRegions[region] = append(a.detectedRegions[region], a.skinMap[currentIndex])
					}
				}
			}
		}
	}

	a.merge(a.detectedRegions, a.mergeRegions)
	a.analyzeRegions()

	return a.result, err
}

func (a *Analyzer) addMerge(from, to int) {
	a.lastFrom = from
	a.lastTo = to

	fromIndex := -1
	toIndex := -1

	for index, region := range a.mergeRegions {
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
			fromRegion := a.mergeRegions[fromIndex]
			toRegion := a.mergeRegions[toIndex]
			region := append(fromRegion, toRegion...)
			a.mergeRegions[fromIndex] = region
			a.mergeRegions = append(a.mergeRegions[:toIndex], a.mergeRegions[toIndex+1:]...)
		}
		return
	}

	if fromIndex == -1 && toIndex == -1 {
		a.mergeRegions = append(a.mergeRegions, []int{from, to})
		return
	}

	if fromIndex != -1 && toIndex == -1 {
		a.mergeRegions[fromIndex] = append(a.mergeRegions[fromIndex], to)
		return
	}

	if fromIndex == -1 && toIndex != -1 {
		a.mergeRegions[toIndex] = append(a.mergeRegions[toIndex], from)
		return
	}

}

// function for merging detected regions
func (a *Analyzer) merge(detectedRegions SkinMapList, mergeRegions [][]int) {
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
	a.clearRegions(newDetectedRegions)
}

// clean up function
// only push regions which are bigger than a specific amount to the final resul
func (a *Analyzer) clearRegions(detectedRegions SkinMapList) {
	for _, region := range detectedRegions {
		if len(region) > 30 {
			a.SkinRegions = append(a.SkinRegions, region)
		}
	}
}

func (a *Analyzer) analyzeRegions() bool {
	skinRegionLength := len(a.SkinRegions)

	// if there are less than 3 regions
	if skinRegionLength < 3 {
		a.message = fmt.Sprintf("Less than 3 skin regions (%v)", skinRegionLength)
		a.result = false
		return a.result
	}

	// sort the skinRegions
	sort.Sort(a.SkinRegions)
	//sort.Reverse(a.SkinRegions)

	// count total skin pixels
	var totalSkin float64
	for _, region := range a.SkinRegions {
		totalSkin += float64(len(region))
	}

	// check if there are more than 15% skin pixel in the image
	totalSkinParcentage := totalSkin / float64(a.totalPixels) * 100
	if totalSkinParcentage < 15 {
		// if the parcentage lower than 15, it's not nude!
		a.message = fmt.Sprintf("Total skin parcentage lower than 15 (%v%%)", totalSkinParcentage)
		a.result = false
		return a.result
	}

	// check if the largest skin region is less than 35% of the total skin count
	// AND if the second largest region is less than 30% of the total skin count
	// AND if the third largest region is less than 30% of the total skin count
	biggestRegionParcentage := float64(len(a.SkinRegions[0])) / totalSkin * 100
	secondLargeRegionParcentage := float64(len(a.SkinRegions[1])) / totalSkin * 100
	thirdLargesRegionParcentage := float64(len(a.SkinRegions[2])) / totalSkin * 100
	if biggestRegionParcentage < 35 &&
		secondLargeRegionParcentage < 30 &&
		thirdLargesRegionParcentage < 30 {
		a.message = "Less than 35%, 30%, 30% skin in the biggest regions"
		a.result = false
		return a.result
	}

	// check if the number of skin pixels in the largest region is less than 45% of the total skin count
	if biggestRegionParcentage < 45 {
		a.message = fmt.Sprintf("The biggest region contains less than 45%% (%v)", biggestRegionParcentage)
		a.result = false
		return a.result
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
	if skinRegionLength > 60 {
		a.message = fmt.Sprintf("More than 60 skin regions (%v)", skinRegionLength)
		a.result = false
		return a.result
	}

	// otherwise it is nude
	a.result = true
	return a.result
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
