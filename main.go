package main

import (
	"github.com/evnix/natsdash/ds"
	"github.com/rivo/tview"
	"github.com/evnix/natsdash/logger"
)

var app *tview.Application
var pages *tview.Pages
var data *ds.Data

func main() {
	logger.Init()
	defer logger.CloseLogger()
	app = tview.NewApplication()
	pages = tview.NewPages()

	data = &ds.Data{}
	data.Contexts = make([]ds.Context, 0)
	contextPage := NewContextPage(app,data)
	contextFormPage := NewContextFormPage(app, data)
	ServerInfoPage := NewServerInfoPage(app, data)
	natsPage := NewNatsPage(app, data)

	pages.AddPage("natsPage", natsPage, true, false)
	pages.AddPage("contextFormPage", contextFormPage, true, false)
	pages.AddPage("serverInfoPage", ServerInfoPage, true, false)
	pages.AddPage("contexts", contextPage, true, true)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
