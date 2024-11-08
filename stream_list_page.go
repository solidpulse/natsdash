package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
)

type StreamListPage struct {
	*tview.Flex
	Data         *ds.Data
	streamList   *tview.List
	app          *tview.Application
	footerTxt    *tview.TextView
	deleteConfirmStream string    // stream pending deletion
	deleteConfirmTimer  *time.Timer
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
    conn := ctx.Conn

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

	go sp.app.Draw()
}

func (sp *StreamListPage) setupInputCapture() {
	sp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			sp.goBackToContextPage()
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			sp.app.SetFocus(sp.streamList)
		case tcell.KeyRune:
			switch event.Rune() {
			case 'v', 'V':
				if sp.streamList.GetItemCount() > 0 {
					sp.viewStream()
					return nil
				}
			}
		}

		switch event.Rune() {
		case 'a', 'A':
			logger.Info("Add stream action triggered")
			pages.SwitchToPage("streamAddPage")
			_, b := pages.GetFrontPage()
			b.(*StreamAddPage).redraw(&data.CurrCtx)
		case 'e', 'E':
			if sp.streamList.GetItemCount() == 0 {
				sp.notify("No stream selected", 3*time.Second, "error")
				return event
			}
			idx := sp.streamList.GetCurrentItem()
			streamName,_ := sp.streamList.GetItemText(idx)
			logger.Info("Edit stream action triggered for: %s", streamName)
			pages.SwitchToPage("streamAddPage")
			_, b := pages.GetFrontPage()
			addPage := b.(*StreamAddPage)
			addPage.setEditMode(streamName)
			addPage.redraw(&sp.Data.CurrCtx)
		case 'i', 'I':
			logger.Info("Stream info action triggered")
			if sp.streamList.GetItemCount() == 0 {
				sp.notify("No stream selected", 3*time.Second, "error")
				return event
			}
			idx := sp.streamList.GetCurrentItem()
			streamName, _ := sp.streamList.GetItemText(idx)
			logger.Info("Stream info action triggered for: %s", streamName)
			pages.SwitchToPage("streamInfoPage")
			_, b := pages.GetFrontPage()
			infoPage := b.(*StreamInfoPage)
			infoPage.streamName = streamName
			infoPage.redraw(&sp.Data.CurrCtx)
		case 'c', 'C':
			if sp.streamList.GetItemCount() == 0 {
				sp.notify("No stream selected", 3*time.Second, "error")
				return event
			}
			idx := sp.streamList.GetCurrentItem()
			streamName, _ := sp.streamList.GetItemText(idx)
			logger.Info("consumer list action triggered for: %s", streamName)
			pages.SwitchToPage("consumerListPage")
			_, b := pages.GetFrontPage()
			infoPage := b.(*ConsumerListPage)
			infoPage.streamName = streamName
			infoPage.redraw(&sp.Data.CurrCtx)
		case 'd', 'D': 
			if sp.streamList.GetItemCount() == 0 {
				sp.notify("No stream selected", 3*time.Second, "error")
				return event
			}
			idx := sp.streamList.GetCurrentItem()
			streamName, _ := sp.streamList.GetItemText(idx)
			
			if sp.deleteConfirmStream == streamName {
				// Second press - execute delete
				sp.deleteConfirmTimer.Stop()
				sp.deleteConfirmStream = ""
				sp.executeDelete(streamName)
			} else {
				// First press - start confirmation
				logger.Info("Delete stream action triggered for: %s", streamName)
				sp.startDeleteConfirmation(streamName)
			}
			return nil
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

func (sp *StreamListPage) startDeleteConfirmation(streamName string) {
	sp.deleteConfirmStream = streamName
	sp.notify("Press d again within 10 seconds to confirm deletion of '"+streamName+"'", 10*time.Second, "warning")
	
	// Cancel any existing timer
	if sp.deleteConfirmTimer != nil {
		sp.deleteConfirmTimer.Stop()
	}
	
	// Start new timer
	sp.deleteConfirmTimer = time.NewTimer(10 * time.Second)
	go func() {
		<-sp.deleteConfirmTimer.C
		sp.deleteConfirmStream = ""
		sp.notify("Delete confirmation timed out", 3*time.Second, "info")
	}()
}

func (sp *StreamListPage) executeDelete(streamName string) {
	// Get JetStream context
	js, err := sp.Data.CurrCtx.Conn.JetStream()
	if err != nil {
		sp.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Delete the stream
	err = js.DeleteStream(streamName)
	if err != nil {
		sp.notify("Failed to delete stream: "+err.Error(), 3*time.Second, "error")
	} else {
		sp.notify("Stream '"+streamName+"' deleted successfully", 3*time.Second, "info")
		sp.redraw(&sp.Data.CurrCtx)
	}
}

func createStreamListHeaderRow() *tview.Flex {
    container := tview.NewFlex()
    container.SetBorder(false)
    container.
        SetDirection(tview.FlexColumn).
        SetBorderPadding(0, 0, 1, 1)

    col1 := tview.NewFlex()
    col1.SetDirection(tview.FlexRow)
    col1.SetBorder(false)

    col1.AddItem(createTextView("[ESC] Back", tcell.ColorWhite), 0, 1, false)
    col1.AddItem(createTextView(" ", tcell.ColorWhite), 0, 1, false)
    col1.AddItem(createTextView(" ", tcell.ColorWhite), 0, 1, false)

    col2 := tview.NewFlex()
    col2.SetDirection(tview.FlexRow)
    col2.SetBorder(false)

    col2.AddItem(createTextView("[i] Info", tcell.ColorWhite), 0, 1, false)
    col2.AddItem(createTextView("[c] Consumers", tcell.ColorWhite), 0, 1, false)
    col2.AddItem(createTextView("[v] View", tcell.ColorWhite), 0, 1, false)

	col3 := tview.NewFlex()
	col3.SetDirection(tview.FlexRow)
	col3.SetBorder(false)
	col3.AddItem(createTextView("[a] Add", tcell.ColorWhite), 0, 1, false)
	col3.AddItem(createTextView("[e] Edit", tcell.ColorWhite), 0, 1, false)	
	col3.AddItem(createTextView("[d] Delete", tcell.ColorWhite), 0, 1, false)

    container.AddItem(col1, 0, 1, false)
    container.AddItem(col2, 0, 1, false)
	container.AddItem(col3, 0, 1, false)
    container.SetTitle("STREAMS")

    return container
}
func (sp *StreamListPage) viewStream() {
	streamName, _ := sp.streamList.GetItemText(sp.streamList.GetCurrentItem())
	viewPage := pages.GetPage("streamViewPage").(*StreamViewPage)
	viewPage.streamName = streamName
	pages.SwitchToPage("streamViewPage")
	viewPage.redraw(&sp.Data.CurrCtx)
	sp.app.SetFocus(viewPage)
}
