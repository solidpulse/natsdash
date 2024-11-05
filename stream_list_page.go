package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
	"github.com/solidpulse/natsdash/natsutil"
)

type StreamListPage struct {
	*tview.Flex
	Data         *ds.Data
	streamList   *tview.List
	app          *tview.Application
	footerTxt    *tview.TextView
}

func NewStreamListPage(app *tview.Application, data *ds.Data) *StreamListPage {
	sp := &StreamListPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	sp.setupUI()
	sp.setupInputCapture()

	return sp
}

func (sp *StreamListPage) setupUI() {
	// Header setup
	headerRow := createStreamListHeaderRow()
	sp.AddItem(headerRow, 4, 4, false)

	// Stream list setup
	streamListBox := tview.NewFlex()
	streamListBox.SetTitle("Streams").SetBorder(true)
	sp.streamList = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	streamListBox.AddItem(sp.streamList, 0, 20, false)
	streamListBox.SetBorderPadding(0, 0, 1, 1)
	sp.AddItem(streamListBox, 0, 18, false)

	// Footer setup
	footer := tview.NewFlex()
	footer.SetBorder(true)
	footer.SetDirection(tview.FlexRow)
	footer.SetBorderPadding(0, 0, 1, 1)

	sp.footerTxt = createTextView(" -- ", tcell.ColorWhite)
	footer.AddItem(sp.footerTxt, 0, 1, false)
	sp.AddItem(footer, 3, 2, false)
	sp.SetBorderPadding(1, 0, 1, 1)

	// Add dummy streams for testing

}

func (sp *StreamListPage) redraw(ctx *ds.Context) {
    sp.streamList.Clear()
    
    // Connect to NATS
    conn, err := natsutil.Connect(&ctx.CtxData)
    if err != nil {
        logger.Error("Failed to connect to NATS server: %v", err)
        sp.notify("Failed to connect to NATS server", 3*time.Second, "error")
        return
    }
    defer conn.Close() // Ensure connection is closed when we're done

    // Get JetStream context
    js, err := conn.JetStream()
    if err != nil {
        logger.Error("Failed to get JetStream context: %v", err)
        sp.notify("Failed to get JetStream context", 3*time.Second, "error")
        return
    }

    // List all streams
    streams := make([]string, 0)
    for stream := range js.StreamNames() {
        streams = append(streams, stream)
    }

    // Add streams to the list
    for _, stream := range streams {
        sp.streamList.AddItem(stream, "", 0, nil)
    }

    if len(streams) == 0 {
        sp.notify("No streams found", 3*time.Second, "info")
    }
}

func (sp *StreamListPage) setupInputCapture() {
	sp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			sp.goBackToContextPage()
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			sp.app.SetFocus(sp.streamList)
		}

		switch event.Rune() {
		case 'a', 'A':
			logger.Info("Add stream action triggered")
			pages.SwitchToPage("streamAddPage")
			_, b := pages.GetFrontPage()
			b.(*StreamAddPage).redraw(&data.CurrCtx)
		case 'e', 'E':
			logger.Info("Edit stream action triggered")
			sp.notify("Edit stream functionality coming soon...", 3*time.Second, "info")
		case 'i', 'I':
			logger.Info("Stream info action triggered")
			sp.notify("Stream info functionality coming soon...", 3*time.Second, "info")
		case 'd', 'D':
			logger.Info("Delete stream action triggered")
			sp.notify("Delete stream functionality coming soon...", 3*time.Second, "info")
		}
		return event
	})
}


func (cfp *StreamListPage) goBackToContextPage() {

	pages.SwitchToPage("contexts")
	_, b := pages.GetFrontPage()
	b.(*ContextPage).Redraw()
	cfp.app.SetFocus(b) // Add this line
}


func (sp *StreamListPage) notify(message string, duration time.Duration, logLevel string) {
	sp.footerTxt.SetText(message)
	sp.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		sp.footerTxt.SetText("")
		sp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}

func createStreamListHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(0, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createTextView("[ESC] Back", tcell.ColorWhite), 0, 1, false)
	headerRow1.AddItem(createTextView("[a] Add", tcell.ColorWhite), 0, 1, false)
	headerRow1.AddItem(createTextView("[e] Edit", tcell.ColorWhite), 0, 1, false)

	headerRow2 := tview.NewFlex()
	headerRow2.SetDirection(tview.FlexRow)
	headerRow2.SetBorder(false)

	headerRow2.AddItem(createTextView("[i] Info", tcell.ColorWhite), 0, 1, false)
	headerRow2.AddItem(createTextView("[DEL] Delete", tcell.ColorWhite), 0, 1, false)
	headerRow2.AddItem(createTextView("", tcell.ColorWhite), 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.AddItem(headerRow2, 0, 1, false)
	headerRow.SetTitle("STREAMS")

	return headerRow
}
