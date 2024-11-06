package main

import (
	"gopkg.in/yaml.v2"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
)

type StreamInfoPage struct {
	*tview.Flex
	app        *tview.Application
	Data       *ds.Data
	textArea   *tview.TextArea
	footerTxt  *tview.TextView
	isEdit     bool
	streamName string
}

func NewStreamInfoPage(app *tview.Application, data *ds.Data) *StreamInfoPage {
	sap := &StreamInfoPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	sap.setupUI()
	sap.setupInputCapture()
	return sap
}

func (sap *StreamInfoPage) setEditMode(name string) {
	sap.isEdit = true
	sap.streamName = name
}

func (sap *StreamInfoPage) setupUI() {
	// Header
	headerRow := tview.NewFlex()
	headerRow.SetDirection(tview.FlexColumn)
	headerRow.SetBorderPadding(1, 0, 1, 1)

	headerRow.AddItem(createTextView("[ESC] Back", tcell.ColorWhite), 0, 1, false)
	headerRow.SetTitle("STREAM INFO")
	sap.AddItem(headerRow, 3, 1, false)

	// Text area for YAML
	sap.textArea = tview.NewTextArea()
	sap.textArea.SetBorder(true)
	sap.textArea.SetDisabled(true)
	sap.textArea.SetTitle("Stream Info")
	sap.AddItem(sap.textArea, 0, 1, true)

	// Footer
	footer := tview.NewFlex()
	footer.SetBorder(true)
	sap.footerTxt = createTextView("", tcell.ColorWhite)
	footer.AddItem(sap.footerTxt, 0, 1, false)
	sap.AddItem(footer, 3, 1, false)

	userFriendlyJSON5 := ``
	sap.textArea.SetText(userFriendlyJSON5, false)
}

func (sap *StreamInfoPage) setupInputCapture() {
	sap.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			sap.notify("Loading......", 3*time.Second, "info")
			sap.goBack()
			return nil
		}
		return event
	})
}

func (sap *StreamInfoPage) goBack() {
	pages.SwitchToPage("streamListPage")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(&sap.Data.CurrCtx)
	sap.app.SetFocus(b)
}

func (sap *StreamInfoPage) redraw(ctx *ds.Context) {
	if !sap.isEdit {
		return
	}

	// Connect to NATS
	conn := ctx.Conn

	// Get JetStream context
	js, err := conn.JetStream()
	if err != nil {
		sap.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Get stream info
	stream, err := js.StreamInfo(sap.streamName)
	if err != nil {
		sap.notify("Failed to get stream info: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(stream.Config)
	if err != nil {
		sap.notify("Failed to convert to YAML: "+err.Error(), 3*time.Second, "error")
		return
	}
	
	yamlTxt := string(yamlBytes)
	sap.textArea.SetText(yamlTxt, true)
	go sap.app.Draw()
    
}

func (sap *StreamInfoPage) notify(message string, duration time.Duration, logLevel string) {
	sap.footerTxt.SetText(message)
	sap.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		sap.footerTxt.SetText("")
		sap.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}
