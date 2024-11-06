package main

import (
	"encoding/json"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
)

type ConsumerAddPage struct {
	*tview.Flex
	Data           *ds.Data
	consumerList   *tview.List
	app            *tview.Application
	txtArea       *tview.TextArea
	footerTxt      *tview.TextView
	streamName       string
	consumerName     string
	editMode       bool
}

func NewConsumerAddPage(app *tview.Application, data *ds.Data) *ConsumerAddPage {
	cp := &ConsumerAddPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	// Create header
	headerRow1 := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerRow1.AddItem(createTextView("Consumer List", tcell.ColorYellow), 0, 1, false)

	headerRow2 := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerRow2.AddItem(createTextView("[ESC] Back [alt+Enter] Save ", tcell.ColorWhite), 0, 1, false)

	// txtarea
	cp.txtArea = tview.NewTextArea()
	cp.txtArea.SetBorder(true)
	cp.txtArea.SetTitle("Consumer Configuration (JSON5)")
	cp.txtArea.SetBorder(true)
	cp.txtArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cp.goBack()
			return nil
		}
		return event
	})
	cp.AddItem(cp.txtArea, 0, 1, true)

	// Footer
	footer := tview.NewFlex()
	footer.SetBorder(true)
	footer.SetDirection(tview.FlexRow)
	footer.AddItem(cp.footerTxt, 0, 1, false)
	cp.AddItem(footer, 3, 1, false)

	cp.SetBorderPadding(1, 1, 1, 1)

	return cp
}

func (cp *ConsumerAddPage) redraw(ctx *ds.Context) {
	//get the consumer details if edit
	if cp.editMode {
		// Connect to NATS
		conn := ctx.Conn

		// Get JetStream context
		js, err := conn.JetStream()
		if err != nil {
			cp.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
			return
		}

		// Get stream info
		stream, err := js.ConsumerInfo(cp.streamName, cp.consumerName)
		if err != nil {
			cp.notify("Failed to get consumer info: "+err.Error(), 3*time.Second, "error")
			return
		}

		// Convert to JSON
		jsonBytes, err := json.Marshal(stream.Config)
		if err != nil {
			cp.notify("Failed to convert to JSON: "+err.Error(), 3*time.Second, "error")
			return
		}

		jsonStr := string(jsonBytes)
		cp.txtArea.SetText(jsonStr, false)
		go cp.app.Draw()
	}else{
		cp.txtArea.SetText("", false)
	}
}
	
func (cp *ConsumerAddPage) goBack() {
	pages.SwitchToPage("streamListPage")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(&cp.Data.CurrCtx)
	cp.app.SetFocus(b) // Add this line
}

func (cp *ConsumerAddPage) notify(message string, duration time.Duration, logLevel string) {
	cp.footerTxt.SetText(message)
	cp.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		cp.footerTxt.SetText("")
		cp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}