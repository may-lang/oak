package render

import (
	"image"

	"github.com/oakmound/oak/dlog"
	"github.com/oakmound/oak/oakerr"
)

// LoadSheet loads a file in some directory with sheets of (w,h) sized sprites,
// where there is pad pixels of vertical/horizontal pad between each sprite.
// This will blow away any cached sheet with the same fileName.
func LoadSheet(directory, fileName string, w, h, pad int) (*Sheet, error) {

	if w <= 0 {
		dlog.Error("Bad dimensions given to load sheet")
		return nil, oakerr.InvalidInput{InputName: "w"}
	}
	if h <= 0 {
		dlog.Error("Bad dimensions given to load sheet")
		return nil, oakerr.InvalidInput{InputName: "h"}
	}
	if pad < 0 {
		dlog.Error("Bad pad given to load sheet")
		return nil, oakerr.InvalidInput{InputName: "pad"}
	}

	var rgba *image.RGBA
	var ok bool
	var err error

	imageLock.RLock()
	rgba, ok = loadedImages[fileName]
	imageLock.RUnlock()

	if !ok {
		dlog.Verb("Missing file in loaded images: ", fileName)
		rgba, err = loadSprite(directory, fileName)
		if err != nil {
			return nil, err
		}
	}

	dlog.Verb("Loading sheet: ", fileName)

	bounds := rgba.Bounds()

	sheetW := bounds.Max.X / w
	remainderW := bounds.Max.X % w
	sheetH := bounds.Max.Y / h
	remainderH := bounds.Max.Y % h

	var widthBuffers, heightBuffers int
	if pad != 0 {
		widthBuffers = remainderW / pad
		heightBuffers = remainderH / pad
	} else {
		widthBuffers = sheetW - 1
		heightBuffers = sheetH - 1
	}

	if sheetW < 1 || sheetH < 1 ||
		widthBuffers != sheetW-1 ||
		heightBuffers != sheetH-1 {
		dlog.Error("Bad dimensions given to load sheet")
		return nil, oakerr.InvalidInput{InputName: "w,h"}
	}

	sheet := make(Sheet, sheetW)
	i := 0
	for x := 0; x < bounds.Max.X; x += (w + pad) {
		sheet[i] = make([]*image.RGBA, sheetH)
		j := 0
		for y := 0; y < bounds.Max.Y; y += (h + pad) {
			sheet[i][j] = subImage(rgba, x, y, w, h)
			j++
		}
		i++
	}

	dlog.Verb("Loaded sheet into map")
	sheetLock.Lock()
	defer sheetLock.Unlock()
	loadedSheets[fileName] = &sheet

	return loadedSheets[fileName], nil
}

// GetSheet tries to find the given file in the set of loaded sheets.
// If SheetIsLoaded(filename) is not true, this returns an error.
// Otherwise it will return the sheet as a 2d array of sprites
func GetSheet(fileName string) (*Sheet, error) {
	sheetLock.RLock()
	dlog.Verb(loadedSheets, fileName, loadedSheets[fileName])
	sh, ok := loadedSheets[fileName]
	if !ok {
		return nil, oakerr.NotFound{InputName: fileName}
	}
	sheetLock.RUnlock()
	return sh, nil
}

// LoadSheetSequence loads a sheet and then calls LoadSequence on that sheet
func LoadSheetSequence(fileName string, w, h, pad int, fps float64, frames ...int) (*Sequence, error) {
	sheet, err := LoadSheet(dir, fileName, w, h, pad)
	if err != nil {
		return nil, err
	}
	return NewSheetSequence(sheet, fps, frames...)
}

// SheetIsLoaded returns whether when LoadSheet is called, a cached sheet will
// be used, or if false that a new file will attempt to be loaded and stored
func SheetIsLoaded(fileName string) bool {
	sheetLock.RLock()
	_, ok := loadedSheets[fileName]
	sheetLock.RUnlock()
	return ok
}
