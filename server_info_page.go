package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
)

type ServerInfoPage struct {
	*tview.Flex
	Data *ds.Data
	form *tview.Form
	app  *tview.Application // Add this line
}

func NewServerInfoPage(app *tview.Application, data *ds.Data) *ServerInfoPage {
	cfp := &ServerInfoPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app, // Add this line
	}
	cfp.Data = data
	cfp.setupUI()
	cfp.setupInputCapture()

	return cfp
}

func (cfp *ServerInfoPage) setupUI() {
	// Header setup
	headerRow := createServerInfoHeaderRow()
	cfp.AddItem(headerRow, 0, 4, false)

	txtView := createTextView("Coming soon...", tcell.ColorWhite)
	cfp.AddItem(txtView, 0, 8, false)


	// // Footer setup
	// footer := tview.NewFlex().SetBorder(true)
	// cfp.AddItem(footer, 0, 1, false)

	cfp.SetBorderPadding(1, 1, 1, 1)
}

func (cfp *ServerInfoPage) redraw(ctx *ds.Context) {

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

func (cfp *ServerInfoPage) setupInputCapture() {
	cfp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cfp.goBackToContextPage()
			return nil
		}
		return event
	})
}

func (cfp *ServerInfoPage) goBackToContextPage() {
	pages.SwitchToPage("contexts")
	_, b := pages.GetFrontPage()
	b.(*ContextPage).Redraw()
	cfp.app.SetFocus(b) // Add this line
}

func createServerInfoHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createTextView("[Esc] Back", tcell.ColorWhite), 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.SetTitle("NATS-DASH")

	return headerRow
}
