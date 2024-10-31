package main

import (
	"github.com/evnix/natsdash/ds"
	"github.com/rivo/tview"
)

var app *tview.Application
var pages *tview.Pages
var data *ds.Data

func main() {
	app = tview.NewApplication()
	pages = tview.NewPages()

	data = &ds.Data{}
	data.Contexts = make([]ds.Context, 0)
	data.LoadFromFile()
	contextPage := NewContextPage(data)
	contextPage.Redraw()
	contextFormPage := NewContextFormPage(data)

	pages.AddPage("contexts", contextPage, true, true)
	pages.AddPage("contextFormPage", contextFormPage, true, false)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
