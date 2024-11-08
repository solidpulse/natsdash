package main

import (
	"io"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nats-io/nats.go"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
)

type StreamViewPage struct {
	*tview.Flex
	Data          *ds.Data
	app           *tview.Application
	streamName    string
	filterSubject *tview.InputField
	logView       *tview.TextView
	subjectName   *tview.InputField
	txtArea       *tview.TextArea
	footerTxt     *tview.TextView
	consumer      *nats.Consumer
}

func NewStreamViewPage(app *tview.Application, data *ds.Data) *StreamViewPage {
	svp := &StreamViewPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}
	svp.setupUI()
	svp.setupInputCapture()
	return svp
}

func (svp *StreamViewPage) setupUI() {
	// Header setup
	headerRow := createStreamViewHeaderRow()
	svp.AddItem(headerRow, 2, 6, false)

	// Filter subject field
	svp.filterSubject = tview.NewInputField()
	svp.filterSubject.SetLabel("Filter Subject: ")
	svp.filterSubject.SetBorder(true)
	svp.filterSubject.SetBorderPadding(0, 0, 1, 1)
	svp.filterSubject.SetDoneFunc(func(key tcell.Key) {
		svp.updateConsumerFilter()
		svp.app.SetFocus(svp.logView)
	})
	svp.AddItem(svp.filterSubject, 3, 6, false)

	// Log view for messages
	svp.logView = tview.NewTextView()
	svp.logView.SetBorder(true)
	svp.logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			svp.app.SetFocus(svp.subjectName)
			return nil
		}
		return event
	})
	svp.AddItem(svp.logView, 0, 50, false)

	// Target subject field for publishing
	svp.subjectName = tview.NewInputField()
	svp.subjectName.SetLabel("Target Subject: ")
	svp.subjectName.SetBorder(true)
	svp.subjectName.SetDoneFunc(func(key tcell.Key) {
		svp.app.SetFocus(svp.txtArea)
	})
	svp.AddItem(svp.subjectName, 3, 6, false)

	// Message text area
	svp.txtArea = tview.NewTextArea()
	svp.txtArea.SetPlaceholder("Message...")
	svp.txtArea.SetBorder(true)
	svp.txtArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			svp.app.SetFocus(svp.filterSubject)
			return nil
		} else if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			svp.publishMessage()
			return nil
		}
		return event
	})
	svp.AddItem(svp.txtArea, 0, 8, false)

	// Footer for notifications
	svp.footerTxt = tview.NewTextView()
	svp.footerTxt.SetBorder(true)
	svp.AddItem(svp.footerTxt, 3, 2, false)

	svp.SetBorderPadding(0, 0, 1, 1)
}

func (svp *StreamViewPage) setupInputCapture() {
	svp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			svp.goBack()
			return nil
		case tcell.KeyLeft:
			svp.fetchPreviousMessage()
			return nil
		case tcell.KeyRight:
			svp.fetchNextMessage()
			return nil
		}
		return event
	})
}

func (svp *StreamViewPage) redraw(ctx *ds.Context) {
	svp.logView.Clear()
	svp.createTemporaryConsumer()
	svp.app.SetFocus(svp.filterSubject)
}

func (svp *StreamViewPage) createTemporaryConsumer() {
	js, err := svp.Data.CurrCtx.Conn.JetStream()
	if err != nil {
		svp.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Create a unique name for temporary consumer
	tempName := "TEMP_VIEW_" + time.Now().Format("20060102150405")

	// Create consumer config
	cfg := &nats.ConsumerConfig{
		Durable:       tempName,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: svp.filterSubject.GetText(),
	}

	consumer, err := js.AddConsumer(svp.streamName, cfg)
	if err != nil {
		svp.notify("Failed to create consumer: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Clean up previous consumer if exists
	if svp.consumer != nil {
		js.DeleteConsumer(svp.streamName, svp.consumer.CachedInfo().Name)
	}

	svp.consumer = consumer
}

func (svp *StreamViewPage) updateConsumerFilter() {
	if svp.consumer != nil {
		svp.createTemporaryConsumer() // Recreate with new filter
	}
}

func (svp *StreamViewPage) fetchNextMessage() {
	if svp.consumer == nil {
		return
	}

	msg, err := svp.consumer.NextMsg(time.Second)
	if err != nil {
		if err != nats.ErrTimeout {
			svp.notify("Failed to fetch message: "+err.Error(), 3*time.Second, "error")
		}
		return
	}

	svp.displayMessage(msg)
	msg.Ack()
}

func (svp *StreamViewPage) fetchPreviousMessage() {
	// Implementation depends on NATS server version and capabilities
	svp.notify("Previous message functionality not implemented", 3*time.Second, "warning")
}

func (svp *StreamViewPage) publishMessage() {
	js, err := svp.Data.CurrCtx.Conn.JetStream()
	if err != nil {
		svp.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	subject := svp.subjectName.GetText()
	message := svp.txtArea.GetText()

	_, err = js.Publish(subject, []byte(message))
	if err != nil {
		svp.notify("Failed to publish message: "+err.Error(), 3*time.Second, "error")
		return
	}

	svp.notify("Message published successfully", 3*time.Second, "info")
	svp.txtArea.SetText("", true)
}

func (svp *StreamViewPage) displayMessage(msg *nats.Msg) {
	timestamp := time.Now().Format("15:04:05.00000")
	text := timestamp + " [" + msg.Subject + "] " + string(msg.Data) + "\n"
	svp.logView.Write([]byte(text))
	svp.logView.ScrollToEnd()
}

func (svp *StreamViewPage) notify(message string, duration time.Duration, logLevel string) {
	svp.footerTxt.SetText(message)
	svp.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		svp.footerTxt.SetText("")
		svp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}

func (svp *StreamViewPage) goBack() {
	if svp.consumer != nil {
		js, _ := svp.Data.CurrCtx.Conn.JetStream()
		js.DeleteConsumer(svp.streamName, svp.consumer.CachedInfo().Name)
	}
	pages.SwitchToPage("streamListPage")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(&svp.Data.CurrCtx)
	svp.app.SetFocus(b)
}

func createStreamViewHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createTextView("[Esc] Back | [Tab] Next Field | [←/→] Navigate Messages | [Alt+Enter] Send", tcell.ColorWhite), 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.SetTitle("Stream View")

	return headerRow
}
