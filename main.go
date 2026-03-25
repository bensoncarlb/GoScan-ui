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

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("TabContainer Widget")

	status := buildStatus()

	docs := buildDocuments()

	mgmt := buildManagement()

	tabs := container.NewAppTabs(
		container.NewTabItem("Status", status),
		container.NewTabItem("Documents", docs),
		container.NewTabItem("Management", mgmt),
	)

	//tabs.Append(container.NewTabItemWithIcon("Home", theme.HomeIcon(), widget.NewLabel("Home tab")))

	tabs.SetTabLocation(container.TabLocationLeading)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}

func buildStatus() fyne.CanvasObject {
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

	return container.NewAdaptiveGrid(2, statusIcon, statusMessage)
}

func buildDocuments() fyne.CanvasObject {
	res, err := http.Get("http://localhost:8090/getdoctypes")

	if err != nil {
		return container.NewHBox(widget.NewLabel("Failed to connect:" + err.Error()))
	}

	rsp := structs.RspGetDocumentTypes{}

	err = json.NewDecoder(res.Body).Decode(&rsp)

	if err != nil {
		panic(err)
	}

	return widget.NewList(
		func() int {
			return len(rsp.DocumentTypes)
		},
		func() fyne.CanvasObject {
			return container.NewPadded(
				widget.NewLabel("Will be replaced"),
				widget.NewButton("Do Something", nil),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[0].(*widget.Label).SetText(rsp.DocumentTypes[id].Title)

			// new part
			item.(*fyne.Container).Objects[1].(*widget.Button).OnTapped = func() {
				deleteDocType(rsp.DocumentTypes[id].Identifier)
				//TODOD remove record from list
			}
		},
	)
}

func buildManagement() fyne.CanvasObject {
	mgmt := widget.NewLabel("Mgmt")

	return mgmt
}

func deleteDocType(docIdentifier string) {
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
