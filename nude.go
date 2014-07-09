package nude

import (
	"fmt"
	"image"
	"math"
	"path/filepath"
	"sort"
)

func IsNude(imageFilePath string) (bool, error) {
	return IsFileNude(imageFilePath)
}

func IsFileNude(imageFilePath string) (bool, error) {
	path, err := filepath.Abs(imageFilePath)
	if err != nil {
		return false, err
	}

	img, err := decodeImage(path)
	if err != nil {
		return false, err
	}

	return IsImageNude(img)
}

func IsImageNude(img image.Image) (bool, error) {
	d:= NewDetector(img)
	return d.Parse()
}

type Detector struct {
	image           image.Image
	width           int
	height          int
	totalPixels     int
	pixels          Region
	SkinRegions     Regions
	detectedRegions Regions
	mergeRegions    [][]int
	lastFrom        int
	lastTo          int
	message         string
	result          bool
}

func NewDetector(img image.Image) *Detector {
	d := &Detector{image: img }
	return d
}

func (d *Detector) Parse() (result bool, err error) {
	img := d.image
	bounds := img.Bounds()
	d.image = img
	d.width = bounds.Size().X
	d.height = bounds.Size().Y
	d.lastFrom = -1
	d.lastTo = -1
	d.totalPixels = d.width * d.height

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		width := bounds.Size().X
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := d.image.At(x, y).RGBA()
			normR := r / 256
			normG := g / 256
			normB := b / 256
			currentIndex := x + y*width
			nextIndex := currentIndex + 1

			isSkin, v := classifySkin(normR, normG, normB)
			if !isSkin {
				d.pixels = append(d.pixels, &Pixel{currentIndex, false, 0, x, y, false, v})
			} else {
				d.pixels = append(d.pixels, &Pixel{currentIndex, true, 0, x, y, false, v})

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
					skin := d.pixels[checkIndex]
					if skin != nil && skin.isSkin {
						if skin.region != region &&
							region != -1 &&
							d.lastFrom != region &&
							d.lastTo != skin.region {
							d.addMerge(region, skin.region)
						}
						region = d.pixels[checkIndex].region
						checker = true
					}
				}

				if !checker {
					d.pixels[currentIndex].region = len(d.detectedRegions)
					d.detectedRegions = append(d.detectedRegions, Region{d.pixels[currentIndex]})
					continue
				} else {
					if region > -1 {
						if len(d.detectedRegions) >= region {
							d.detectedRegions = append(d.detectedRegions, Region{})
						}
						d.pixels[currentIndex].region = region
						d.detectedRegions[region] = append(d.detectedRegions[region], d.pixels[currentIndex])
					}
				}
			}
		}
	}

	d.merge(d.detectedRegions, d.mergeRegions)
	d.analyzeRegions()

	return d.result, err
}

func (d *Detector) addMerge(from, to int) {
	d.lastFrom = from
	d.lastTo = to

	fromIndex := -1
	toIndex := -1

	for index, region := range d.mergeRegions {
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
			fromRegion := d.mergeRegions[fromIndex]
			toRegion := d.mergeRegions[toIndex]
			region := append(fromRegion, toRegion...)
			d.mergeRegions[fromIndex] = region
			d.mergeRegions = append(d.mergeRegions[:toIndex], d.mergeRegions[toIndex+1:]...)
		}
		return
	}

	if fromIndex == -1 && toIndex == -1 {
		d.mergeRegions = append(d.mergeRegions, []int{from, to})
		return
	}

	if fromIndex != -1 && toIndex == -1 {
		d.mergeRegions[fromIndex] = append(d.mergeRegions[fromIndex], to)
		return
	}

	if fromIndex == -1 && toIndex != -1 {
		d.mergeRegions[toIndex] = append(d.mergeRegions[toIndex], from)
		return
	}

}

// function for merging detected regions
func (d *Detector) merge(detectedRegions Regions, mergeRegions [][]int) {
	var newDetectedRegions Regions

	// merging detected regions
	for index, region := range mergeRegions {
		if len(newDetectedRegions) >= index {
			newDetectedRegions = append(newDetectedRegions, Region{})
		}
		for _, r := range region {
			newDetectedRegions[index] = append(newDetectedRegions[index], detectedRegions[r]...)
			detectedRegions[r] = Region{}
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
	d.clearRegions(newDetectedRegions)
}

// clean up function
// only push regions which are bigger than a specific amount to the final resul
func (d *Detector) clearRegions(detectedRegions Regions) {
	for _, region := range detectedRegions {
		if len(region) > 30 {
			d.SkinRegions = append(d.SkinRegions, region)
		}
	}
}

func (d *Detector) analyzeRegions() bool {
	skinRegionLength := len(d.SkinRegions)

	// if there are less than 3 regions
	if skinRegionLength < 3 {
		d.message = fmt.Sprintf("Less than 3 skin regions (%v)", skinRegionLength)
		d.result = false
		return d.result
	}

	// sort the skinRegions
	sort.Sort(sort.Reverse(d.SkinRegions))

	// count total skin pixels
	totalSkinPixels := float64(d.SkinRegions.totalPixels())

	// check if there are more than 15% skin pixel in the image
	totalSkinParcentage := totalSkinPixels / float64(d.totalPixels) * 100
	if totalSkinParcentage < 15 {
		// if the parcentage lower than 15, it's not nude!
		d.message = fmt.Sprintf("Total skin parcentage lower than 15 (%v%%)", totalSkinParcentage)
		d.result = false
		return d.result
	}

	// check if the largest skin region is less than 35% of the total skin count
	// AND if the second largest region is less than 30% of the total skin count
	// AND if the third largest region is less than 30% of the total skin count
	biggestRegionParcentage := float64(len(d.SkinRegions[0])) / totalSkinPixels * 100
	secondLargeRegionParcentage := float64(len(d.SkinRegions[1])) / totalSkinPixels * 100
	thirdLargesRegionParcentage := float64(len(d.SkinRegions[2])) / totalSkinPixels * 100
	if biggestRegionParcentage < 35 &&
		secondLargeRegionParcentage < 30 &&
		thirdLargesRegionParcentage < 30 {
		d.message = "Less than 35%, 30%, 30% skin in the biggest regions"
		d.result = false
		return d.result
	}

	// check if the number of skin pixels in the largest region is less than 45% of the total skin count
	if biggestRegionParcentage < 45 {
		d.message = fmt.Sprintf("The biggest region contains less than 45%% (%v)", biggestRegionParcentage)
		d.result = false
		return d.result
	}

	// check if the total skin count is less than 30% of the total number of pixels
	// AND the number of skin pixels within the bounding polygon is less than 55% of the size of the polygon
	// if this condition is true, it's not nude.
	if totalSkinParcentage < 30 {
		for i, region := range d.SkinRegions {
			skinRate := region.skinRateInBoundingPolygon()
			//fmt.Printf("skinRate[%v] = %v\n", i, skinRate)
			if skinRate < 0.55 {
				d.message = fmt.Sprintf("region[%d].skinRate(%v) < 0.55", i, skinRate)
				d.result = false
				return d.result
			}
		}
	}

	// if there are more than 60 skin regions and the average intensity within the polygon is less than 0.25
	// the image is not nude
	averageIntensity := d.SkinRegions.averageIntensity()
	if skinRegionLength > 60 && averageIntensity < 0.25 {
		d.message = fmt.Sprintf("More than 60 skin regions(%v) and averageIntensity(%v) < 0.25", skinRegionLength, averageIntensity)
		d.result = false
		return d.result
	}

	// otherwise it is nude
	d.result = true
	return d.result
}

func (d *Detector) String() string {
	return fmt.Sprintf("#<nude.Detector result=%t, message=%s>", d.result, d.message)
}

// A Survey on Pixel-Based Skin Color Detection Techniques
func classifySkin(r, g, b uint32) (bool, float64) {
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

	h, s, v := toHsv(r, g, b)
	hsvClassifier := h > 0 &&
		h < 35 &&
		s > 0.23 &&
		s < 0.68

	// ycc doesnt work

	result := rgbClassifier || normalizedRgbClassifier || hsvClassifier
	return result, v
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
