package main

import (
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
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
	contextPage := NewContextPage(app, data)
	contextFormPage := NewContextFormPage(app, data)
	ServerInfoPage := NewServerInfoPage(app, data)
	natsPage := NewNatsPage(app, data)
	streamListPage := NewStreamListPage(app, data)
	StreamAddPage := NewStreamAddPage(app, data)
	StreamInfoPage := NewStreamInfoPage(app, data)
	ConsumerListPage := NewConsumerListPage(app, data)
	ConsumerAddPage := NewConsumerAddPage(app, data)
	ConsumerInfoPage := NewConsumerInfoPage(app, data)
	StreamViewPage := NewStreamViewPage(app, data)

	pages.AddPage("natsPage", natsPage, true, false)
	pages.AddPage("streamListPage", streamListPage, true, false)
	pages.AddPage("consumerListPage", ConsumerListPage, true, false)
	pages.AddPage("consumerAddPage", ConsumerAddPage, true, false)
	pages.AddPage("consumerInfoPage", ConsumerInfoPage, true, false)
	pages.AddPage("streamAddPage", StreamAddPage, true, false)
	pages.AddPage("streamInfoPage", StreamInfoPage, true, false)
	pages.AddPage("streamViewPage", StreamViewPage, true, false)
	pages.AddPage("consumerInfoPage", NewConsumerInfoPage(app, data), true, false)
	pages.AddPage("contextFormPage", contextFormPage, true, false)
	pages.AddPage("serverInfoPage", ServerInfoPage, true, false)
	pages.AddPage("contexts", contextPage, true, true)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
