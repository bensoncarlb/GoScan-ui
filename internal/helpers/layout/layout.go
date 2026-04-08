// layout is a package to structure content into pre-defined layouts
package layout

import (
	"fyne.io/fyne/v2"
)

const defaultDividerWidth = 200

type OneByThree struct {
	left   fyne.CanvasObject
	center fyne.CanvasObject
	right  fyne.CanvasObject
	header fyne.CanvasObject
}
