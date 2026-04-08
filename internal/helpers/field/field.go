// field is a package for drawing and managing meta data of a data field from a Document Type
package field

import (
	"image"

	"github.com/bensoncarlb/GoScan/structs"
)

type Field struct {
	DocType       structs.DocumentType
	OverlayRegion image.Rectangle
}
