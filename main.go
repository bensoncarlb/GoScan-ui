package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"github.com/bensoncarlb/GoScan/structs"

	"fyne.io/fyne/v2/widget"
)

type goScanUI struct {
	app fyne.App
}

func main() {
	ui := goScanUI{}

	ui.app = app.New()
	myWindow := ui.app.NewWindow("TabContainer Widget")

	tabs := container.NewAppTabs(
		container.NewTabItem("Status", ui.buildStatus()),
		container.NewTabItem("Documents", ui.buildProcessedDocs()),
		container.NewTabItem("Types", ui.buildDocumentTypes()),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
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

	return widget.NewList(
		func() int {
			return len(rsp.DocumentTypes)
		},
		func() fyne.CanvasObject {
			return container.NewPadded(
				widget.NewLabel("Will be replaced"),
				widget.NewButton("Delete", nil),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[0].(*widget.Label).SetText(rsp.DocumentTypes[id].Title)

			// new part
			item.(*fyne.Container).Objects[1].(*widget.Button).OnTapped = func() {
				a.deleteDocType(rsp.DocumentTypes[id].Identifier)
				//TODOD remove record from list
			}
		},
	)
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

	win := a.app.NewWindow("Document: " + itemName)

	win.SetContent(widget.NewLabel(fmt.Sprintf("%v", doc.Fields)))
	win.Show()
}
