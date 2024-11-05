package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nats-io/nats.go"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"gopkg.in/yaml.v3"
)

type StreamAddPage struct {
	*tview.Flex
	app       *tview.Application
	Data      *ds.Data
	textArea  *tview.TextArea
	footerTxt *tview.TextView
}

func NewStreamAddPage(app *tview.Application, data *ds.Data) *StreamAddPage {
	sap := &StreamAddPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	sap.setupUI()
	sap.setupInputCapture()
	return sap
}

func (sap *StreamAddPage) setupUI() {
	// Header
	headerRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		SetBorderPadding(0, 0, 1, 1)

	headerRow.AddItem(createTextView("[ESC] Back", tcell.ColorWhite), 0, 1, false)
	headerRow.AddItem(createTextView("[Alt+Enter] Save", tcell.ColorWhite), 0, 1, false)
	headerRow.SetTitle("ADD STREAM")
	sap.AddItem(headerRow, 3, 1, false)

	// Text area for YAML
	sap.textArea = tview.NewTextArea()
	sap.textArea.SetBorder(true)
	sap.textArea.SetTitle("Stream Configuration (YAML)")
	sap.AddItem(sap.textArea, 0, 1, true)

	// Footer
	footer := tview.NewFlex().SetBorder(true)
	sap.footerTxt = createTextView("", tcell.ColorWhite)
	footer.AddItem(sap.footerTxt, 0, 1, false)
	sap.AddItem(footer, 3, 1, false)

	// Set default config
	defaultConfig := nats.StreamConfig{
		Name:              "my_stream",
		Description:       "My Stream Description",
		Subjects:         []string{"my.subject.>"},
		Retention:        nats.LimitsPolicy,
		MaxConsumers:     -1,
		MaxMsgs:          -1,
		MaxBytes:         -1,
		Discard:          nats.DiscardOld,
		MaxAge:           24 * time.Hour,
		MaxMsgsPerSubject: -1,
		MaxMsgSize:       -1,
		Storage:          nats.FileStorage,
		Replicas:         1,
	}

	yamlBytes, _ := yaml.Marshal(defaultConfig)
	sap.textArea.SetText(string(yamlBytes), true)
}

func (sap *StreamAddPage) setupInputCapture() {
	sap.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			sap.goBack()
			return nil
		}
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			// TODO: Implement save functionality
			sap.notify("Save functionality coming soon...", 3*time.Second, "info")
			return nil
		}
		return event
	})
}

func (sap *StreamAddPage) goBack() {
	pages.SwitchToPage("streams")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(sap.Data.CurrCtx)
	sap.app.SetFocus(b)
}

func (sap *StreamAddPage) notify(message string, duration time.Duration, logLevel string) {
	sap.footerTxt.SetText(message)
	sap.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		sap.footerTxt.SetText("")
		sap.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}

