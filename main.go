package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"net/http"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/x/fyne/wrapper"
	"github.com/bensoncarlb/GoScan/structs"

	"fyne.io/fyne/v2/widget"
)

type goScanUI struct {
	app fyne.App
}

type regionRow struct {
	name string
	x1   float32
	y1   float32
	x2   float32
	y2   float32
}

func main() {
	ui := goScanUI{}

	ui.app = app.New()
	winMain := ui.app.NewWindow("TabContainer Widget")

	tabs := container.NewAppTabs(
		container.NewTabItem("Status", ui.buildStatus()),
		container.NewTabItem("Documents", ui.buildProcessedDocs()),
		container.NewTabItem("Types", ui.buildDocumentTypes()),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	winMain.SetOnDropped(ui.addDocType)
	winMain.SetContent(tabs)
	winMain.ShowAndRun()

}

func (a *goScanUI) buildStatus() fyne.CanvasObject {
	//TODO configurable
	res, err := http.Get("http://localhost:8090/ping")

	statusIcon := widget.NewIcon(theme.ConfirmIcon())
	statusMessage := widget.NewLabel("Successfully connected.")

	if err != nil {
		statusIcon = widget.NewIcon(theme.CancelIcon())
		statusMessage = widget.NewLabel("Failed to connect")
	} else if res.StatusCode != http.StatusOK {
		statusIcon = widget.NewIcon(theme.WarningIcon())
		statusMessage = widget.NewLabel(fmt.Sprintf("Server responded with unexpected code: %d", res.StatusCode))
	}

	return container.NewPadded(statusIcon, statusMessage)
}

func (a *goScanUI) buildProcessedDocs() fyne.CanvasObject {
	rsp, err := http.Get("http://localhost:8090/getitems")

	if err != nil {
		return container.NewHBox(widget.NewLabel("Failed to connect:" + err.Error()))
	}

	items := structs.RspGetItems{}
	json.NewDecoder(rsp.Body).Decode(&items)

	list := widget.NewList(
		func() int {
			return len(items.Items)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(items.Items[i])
		})

	list.OnSelected = func(id widget.ListItemID) {
		a.viewProcessedItem(items.Items[id])
	}

	return list
}

// Retrieve a list of current Document Types and display them
func (a *goScanUI) buildDocumentTypes() fyne.CanvasObject {
	res, err := http.Get("http://localhost:8090/getdoctypes")

	if err != nil {
		return container.NewHBox(widget.NewLabel("Failed to connect:" + err.Error()))
	}

	rsp := structs.RspGetDocumentTypes{}

	err = json.NewDecoder(res.Body).Decode(&rsp)

	if err != nil {
		panic(err)
	}

	list := widget.NewList(
		func() int {
			return len(rsp.DocumentTypes)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(rsp.DocumentTypes[i].Title)
		})

	list.OnSelected = func(id widget.ListItemID) {
		a.deleteDocType(rsp.DocumentTypes[id].Identifier)
	}

	return container.NewPadded(list)
}

// Handle an image being dragged onto the window
// Set up configuring a new Document Type
func (a *goScanUI) addDocType(p fyne.Position, u []fyne.URI) {
	win := a.app.NewWindow("Add a Document Type")

	//Tracking for clicked points on the image
	//Used for creating regions
	var clickRegions []regionRow
	regRow := 0

	var imgScale float32 = 1.0

	//The two entry fields for the Document Type Title and Identifier
	fieldSize := fyne.NewSize(300, 45)
	title := widget.NewEntry()
	ident := widget.NewEntry()

	title.Resize(fieldSize)
	ident.Resize(fieldSize)

	//Create a list to display the recorded regions
	lstRegions := widget.NewList(
		func() int {
			if clickRegions == nil {
				return 0
			}
			return regRow
		},
		func() fyne.CanvasObject {
			return widget.NewEntry()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			//Little backwards, but rather than setting the entry values
			//we save them into the click tracking slice to not overwrite user entry
			clickRegions[i].name = o.(*widget.Entry).Text
		})

	//Allow a user to delete a region by clicking on it from the list
	lstRegions.OnSelected = func(id widget.ListItemID) {
		deleteElem(clickRegions, id)
		regRow -= 1
	}

	//Create the actual form to hold the Title and Identifier entry fields
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Title", Widget: title}, {Text: "Identifier", Widget: ident}},
		OnSubmit: func() {
			lstRegions.Refresh() //One last refresh to get the latest changes
			//Create a slice of regions based on the clicks we tracked
			docRegions := make([]structs.DocumentRegion, len(clickRegions))
			actualRegions := 0

			for _, newReg := range clickRegions {
				if newReg.x1 > 0 && newReg.x2 > 0 {
					docRegions[actualRegions] = structs.DocumentRegion{
						FieldName:   newReg.name,
						RegionTitle: newReg.name,
						Region: image.Rect(
							int(newReg.x1/imgScale),
							int(newReg.y1/imgScale),
							int(newReg.x2/imgScale),
							int(newReg.y2/imgScale))}
					actualRegions += 1
				}
			}

			docNew := structs.DocumentType{
				Title:      title.Text,
				Identifier: ident.Text,
				Regions:    docRegions[:actualRegions]}

			if addDocType(docNew) {
				win.Close()
				return
			}

			println("Add failed.")
		},
	}

	form.Resize(fyne.NewSize(300, 200))

	imgContainer := container.NewWithoutLayout()

	//Load the image the just dropped
	img := canvas.NewImageFromURI(u[0])
	img.FillMode = canvas.ImageFillOriginal
	clicks := 0
	//Use a fyne community plugin to support making the image clickable
	imgTap := wrapper.MakeTappable(img, func(pe *fyne.PointEvent) {
		clicks += 1
		//Don't bother trying to initialize the slice until the user does something
		if clickRegions == nil {
			clickRegions = make([]regionRow, 50) //TODO handle expansion
		}

		switch clicks % 2 {
		case 1:
			//First pair of clicks for a new region
			newRegion := regionRow{x1: pe.Position.X, y1: pe.Position.Y}
			clickRegions[regRow] = newRegion
		case 0:
			//Second click for a new region
			standardizeRegion(&clickRegions[regRow], pe.Position.X, pe.Position.Y)
			rect := canvas.NewRectangle(color.Black)

			rect.FillColor = color.Transparent
			rect.StrokeWidth = 1
			rect.StrokeColor = color.RGBA{R: 255, A: 255}

			rect.Resize(fyne.NewSize(
				clickRegions[regRow].x2-clickRegions[regRow].x1,
				clickRegions[regRow].y2-clickRegions[regRow].y1))

			rect.Move(fyne.NewPos(clickRegions[regRow].x1, clickRegions[regRow].y1))

			imgContainer.Add(rect)
			imgContainer.Refresh()

			lstRegions.Resize(fyne.NewSize(300, float32(regRow+1)*40))
			lstRegions.Refresh()

			regRow += 1
		}
	})

	imgSize := getImageSize(u[0].Path())
	if imgSize.Height > 800 || imgSize.Width > 1000 {
		imgScale = 800 / imgSize.Height
		imgSize.Height = imgSize.Height * imgScale
		imgSize.Width = imgSize.Width * imgScale
	}

	imgTap.Resize(imgSize)
	imgContainer.Add(imgTap)

	fieldContainer := container.NewVBox(form, lstRegions)
	fieldContainer.Resize(fyne.NewSize(300, 800))

	//win.SetContent(container.NewHBox(form, imgContainer, lstRegions))
	win.SetContent(container.NewHBox(fieldContainer, imgContainer))
	win.Show()
}

func deleteElem(s []regionRow, i int) {
	for i < len(s) {
		s[i] = s[i+1]
	}

	s[i] = regionRow{}
}

// Handle if a defined region is inverted
// Second point should always be further down/right than the first
func standardizeRegion(r *regionRow, x float32, y float32) {
	if r.x1 > x {
		r.x1, r.x2 = x, r.x1
	} else {
		r.x2 = x
	}

	if r.y1 > y {
		r.y1, r.y2 = y, r.y1
	} else {
		r.y2 = y
	}
}

func getImageSize(f string) fyne.Size {
	if strings.TrimSpace(f) == "" {
		return fyne.NewSize(100, 100)
	}

	imgFile, err := os.Open(f)

	if err != nil {
		return fyne.NewSize(100, 100)
	}

	img, _, err := image.Decode(imgFile)

	if err != nil {
		return fyne.NewSize(100, 100)
	}

	return fyne.NewSize(float32(img.Bounds().Size().X), float32(img.Bounds().Size().Y))
}

func addDocType(doc structs.DocumentType) bool {
	if !doc.IsValid() {
		return false
	}

	b := bytes.Buffer{}
	err := json.NewEncoder(&b).Encode(doc)

	if err != nil {
		return false
	}

	if res, err := http.Post("http://localhost:8090/adddoctype", "application/json", &b); err != nil {
		return false
	} else if res.StatusCode != http.StatusAccepted {
		return false
	}

	return true
}

func (a *goScanUI) deleteDocType(docIdentifier string) {
	if strings.TrimSpace(docIdentifier) == "" {
		return
	}

	req := structs.ReqDeleteDocumentType{DocumentType: docIdentifier}

	b := bytes.Buffer{}

	err := json.NewEncoder(&b).Encode(req)

	if err != nil {
		panic(err)
	}

	_, err = http.Post("http://localhost:8090/deletedoctype", "application/json", &b)

	if err != nil {
		panic(err)
	}
}

func (a *goScanUI) viewProcessedItem(itemName string) {
	req := structs.ReqRetrieveItem{ItemName: itemName}

	b := bytes.Buffer{}

	err := json.NewEncoder(&b).Encode(req)

	res, err := http.Post("http://localhost:8090/retrieveitem", "application/json", &b)

	if err != nil {
		panic(err)
	}

	doc := structs.RspRetrieveItem{}
	err = json.NewDecoder(res.Body).Decode(&doc)

	if err != nil {
		panic(err)
	}

	winFields := a.app.NewWindow("Document: " + itemName)

	winFields.SetContent(widget.NewLabel(fmt.Sprintf("%v", doc.Fields)))
	winFields.Show()

	winImage := a.app.NewWindow("Image: " + itemName)

	img, _, err := image.Decode(bytes.NewReader(doc.ImgData))

	if err != nil {
		winImage.SetContent(widget.NewLabel("Failed to open image: " + err.Error()))
	}

	contentImage := canvas.NewImageFromImage(img)
	contentImage.FillMode = canvas.ImageFillOriginal

	winImage.SetContent(contentImage)
	winImage.Show()
}
