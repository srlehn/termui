// <Copyright> 2018,2019 Simon Robin Lehn. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package exp

import (
	"fmt"
	"errors"
	"image"

	"github.com/disintegration/imaging"

	. "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var (
	charBoxWidthInPixels, charBoxHeightInPixels   float64
	charBoxWidthColumns,  charBoxHeightRows       int
)


func resizeImage(self *widgets.Image, buf *Buffer) (img image.Image, err error) {
	img = self.Image
	// img = image.NRGBA{}

	// get dimensions //
	// terminal size measured in cells
	imageWidthInColumns := self.Inner.Dx()
	imageHeightInRows   := self.Inner.Dy()

	// terminal size in cells and pixels and calculated terminal character box size in pixels
	var termWidthInColumns, termHeightInRows int
	var charBoxWidthInPixelsTemp, charBoxHeightInPixelsTemp float64
	termWidthInColumns, termHeightInRows, _, _, charBoxWidthInPixelsTemp, charBoxHeightInPixelsTemp, err = getTermSize()
	if err != nil {
		return img, err
	}
	// update if value is more precise
	if termWidthInColumns > charBoxWidthColumns {
		charBoxWidthInPixels = charBoxWidthInPixelsTemp
	}
	if termHeightInRows > charBoxHeightRows {
		charBoxHeightInPixels = charBoxHeightInPixelsTemp
	}
if isTmux {charBoxWidthInPixels, charBoxHeightInPixels = 10, 19}   // mlterm settings (temporary)

	// calculate image size in pixels
	// subtract 1 pixel for small deviations from char box size (float64)
	imageWidthInPixels  := int(float64(imageWidthInColumns) * charBoxWidthInPixels)  - 1
	imageHeightInPixels := int(float64(imageHeightInRows)   * charBoxHeightInPixels) - 1
	if imageWidthInPixels == 0 || imageHeightInPixels == 0 {
		return img, errors.New("could not calculate the image size in pixels")
	}

	// handle only partially displayed image
	// otherwise we get scrolling
	var needsCropX, needsCropY bool
	var imgCroppedWidth, imgCroppedHeight int
	imgCroppedWidth  = imageWidthInPixels
	imgCroppedHeight = imageHeightInPixels
	if self.Max.Y >= int(termHeightInRows) {
		var scrollExtraRows int
		// remove last 2 rows for xterm when cropped vertically to prevent scrolling
		if isXterm {
			scrollExtraRows = 2
		}
		// subtract 1 pixel for small deviations from char box size (float64)
		imgCroppedHeight = int(float64(int(termHeightInRows) - self.Inner.Min.Y - scrollExtraRows) * charBoxHeightInPixels) - 1
		needsCropY = true
	}
	if self.Max.X >= int(termWidthInColumns) {
		var scrollExtraColumns int
		imgCroppedWidth = int(float64(int(termWidthInColumns) - self.Inner.Min.X - scrollExtraColumns) * charBoxWidthInPixels) - 1
		needsCropX = true
	}

	lastImageDimensions := self.GetVisibleArea()
	// this is meant for comparison and for positioning in the ANSI string
	// the Min values are in cells while the Max values are in pixels
	imageDimensions := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: imgCroppedWidth, Y: imgCroppedHeight}}
	self.SetVisibleArea(imageDimensions)
	// print saved ANSI string if image size and position didn't change
	if imageDimensions.Min.X == lastImageDimensions.Min.X && imageDimensions.Min.Y == lastImageDimensions.Min.Y && imageDimensions.Max.X == lastImageDimensions.Max.X && imageDimensions.Max.Y == lastImageDimensions.Max.Y {
		// reuse last encoded image because of unchanged image dimensions
		return img, nil
	}
	lastImageDimensions = imageDimensions

	// resize and crop the image //
	img = imaging.Resize(self.Image, imageWidthInPixels, imageHeightInPixels, imaging.Lanczos)
	if needsCropX || needsCropY {
		img = imaging.Crop(img, image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: imgCroppedWidth, Y: imgCroppedHeight}})
	}
	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		return img, fmt.Errorf("image size in pixels is 0")
	}
	
	return img, err
}