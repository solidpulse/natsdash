package main

import (
	"github.com/rivo/tview"
)

var app *tview.Application
var pages *tview.Pages

func main() {
	app = tview.NewApplication()
	pages = tview.NewPages()

	contextPage := NewContextPage()
	contextFormPage := NewContextFormPage()

	pages.AddPage("contexts", contextPage, true, true)
	pages.AddPage("contextFormPage", contextFormPage, true, false)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
