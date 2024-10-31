package main

import (

	"github.com/evnix/natsdash/ds"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type NatsPage struct {
	*tview.Flex
	Data         *ds.Data
	app          *tview.Application // Add this line
	subjectFilter *tview.InputField
	logView       *tview.TextView
	subjectName   *tview.InputField
	txtArea       *tview.TextArea
}

func NewNatsPage(app *tview.Application, data *ds.Data) *NatsPage {
	cfp := &NatsPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app, // Add this line
	}
	cfp.Data = data
	cfp.setupUI()
	cfp.setupInputCapture()

	return cfp
}

func (cfp *NatsPage) setupUI() {
	// Header setup
	headerRow := createNatsPageHeaderRow()
	cfp.AddItem(headerRow, 0, 6, false)

	// Initialize fields
	cfp.subjectFilter = tview.NewInputField()
	cfp.subjectFilter.SetLabel("Filter Subjects: ")
	cfp.subjectFilter.SetBorder(true)
	cfp.subjectFilter.SetBorderPadding(0, 0, 1, 1)
	cfp.AddItem(cfp.subjectFilter, 0, 6, false)

	cfp.logView = tview.NewTextView()
	cfp.logView.SetTitle(cfp.Data.CurrCtx.LogFilePath)
	cfp.logView.SetBorder(true)
	cfp.AddItem(cfp.logView, 0, 50, false)

	cfp.subjectName = tview.NewInputField()
	cfp.subjectName.SetLabel("Target Subject: ")
	cfp.subjectName.SetBorder(true)
	cfp.AddItem(cfp.subjectName, 0, 6, false)

	cfp.txtArea = tview.NewTextArea()
	cfp.txtArea.SetPlaceholder("Message...")
	cfp.txtArea.SetBorder(true)
	cfp.AddItem(cfp.txtArea, 0, 8, false)
	cfp.SetBorderPadding(0, 0, 1, 1)
}

func (cfp *NatsPage) redraw(ctx *ds.Context) {
	// Update log view title with the current context's log file path
	cfp.logView.SetTitle(ctx.LogFilePath)
}

func (cfp *NatsPage) setupInputCapture() {
	cfp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cfp.goBackToContextPage()
			return nil
		}
		return event
	})
}

func (cfp *NatsPage) goBackToContextPage() {
	pages.SwitchToPage("contexts")
	_, b := pages.GetFrontPage()
	b.(*ContextPage).Redraw()
	cfp.app.SetFocus(b) // Add this line
}

func createNatsPageHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createTextView("[Esc] Back  | [ctrl+Enter] Send | [F2] Filter | [F3] Subject | [F4] Body", tcell.ColorWhite), 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.SetTitle("NATS-DASH")

	return headerRow
}
