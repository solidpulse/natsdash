package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/evnix/natsdash/ds"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type NatsPage struct {
	*tview.Flex
	Data *ds.Data
	app  *tview.Application // Add this line
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
	subjectFilter := tview.NewInputField()
	subjectFilter.SetLabel("Filter Subjects: ")
	subjectFilter.SetBorder(true)
	cfp.AddItem(subjectFilter, 0, 6, false)
	logView := tview.NewTextView()

	logFilePath := path.Join(os.TempDir(), "natsdash", fmt.Sprintf("%s.log", time.Now().Format("20060102-150405")))
	cfp.AddItem(logView, 1, 6, true)

	logView.SetTitle(logFilePath)
	logView.SetBorder(true)
	cfp.AddItem(logView, 0, 50, false)
	subjectName := tview.NewInputField()
	subjectName.SetLabel("Target Subject: ")
	subjectName.SetBorder(true)
	cfp.AddItem(subjectName, 0, 6, false)
	txtArea := tview.NewTextArea()
	txtArea.SetPlaceholder("Message...")
	txtArea.SetBorder(true)
	cfp.AddItem(txtArea, 0, 8, false)
	// // Form setup

	// // Footer setup
	// footer := tview.NewFlex().SetBorder(true)
	// cfp.AddItem(footer, 0, 1, false)

	cfp.SetBorderPadding(0, 0, 1, 1)
}

func (cfp *NatsPage) redraw(ctx *ds.Context) {

	// go func() {
	// 	cfp.form.GetFormItem(0).(*tview.InputField).SetText("Connecting...")
	// }()

	// go func() {
	// 	if ctx.Conn != nil {
	// 		conn, err := natsutil.Connect(ctx.URL)
	// 		if err != nil {
	// 			cfp.form.GetFormItem(0).(*tview.TextView).SetText(err.Error())
	// 		} else {
	// 			ctx.Conn = conn
	// 		}
	// 	} else {
	// 		cfp.form.GetFormItem(0).(*tview.TextView).SetText("Connected")
	// 	}
	// }()

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
