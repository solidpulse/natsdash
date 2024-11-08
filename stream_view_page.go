package main

import (
	"strings"
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
	consumer      *nats.Subscription
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

	// Create ephemeral consumer subscription
	filterSubject := svp.filterSubject.GetText()
	if filterSubject == "" {
		filterSubject = ">" // Subscribe to all subjects if no filter
	}

	// Clean up previous subscription if exists
	if svp.consumer != nil {
		svp.consumer.Unsubscribe()
	}

	// Create new subscription
	sub, err := js.PullSubscribe(filterSubject, tempName, 
		nats.BindStream(svp.streamName),
		nats.AckExplicit())
	if err != nil {
		svp.notify("Failed to create subscription: "+err.Error(), 3*time.Second, "error")
		return
	}

	svp.consumer = sub
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

	msgs, err := svp.consumer.Fetch(1, nats.MaxWait(time.Second))
	if err != nil {
		if err != nats.ErrTimeout {
			svp.notify("Failed to fetch message: "+err.Error(), 3*time.Second, "error")
		}
		return
	}

	if len(msgs) > 0 {
		svp.displayMessage(msgs[0])
		msgs[0].Ack()
	}
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
	if subject == "" {
		svp.notify("Subject cannot be empty", 3*time.Second, "error")
		return
	}

	message := svp.txtArea.GetText()
	if message == "" {
		svp.notify("Message cannot be empty", 3*time.Second, "error")
		return
	}

	// Get stream info to check subjects
	stream, err := js.StreamInfo(svp.streamName)
	if err != nil {
		svp.notify("Failed to get stream info: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Verify subject matches stream's subject filter
	subjectAllowed := false
	for _, s := range stream.Config.Subjects {
		if nats.IsValidSubject(subject) && subjectMatches(s, subject) {
			subjectAllowed = true
			break
		}
	}

	if !subjectAllowed {
		svp.notify("Subject does not match stream's subject filter", 3*time.Second, "error")
		return
	}

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
	logger.Info(message)

	go func() {
		time.Sleep(duration)
		svp.footerTxt.SetText("")
		svp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}

func (svp *StreamViewPage) goBack() {
	if svp.consumer != nil {
		svp.consumer.Unsubscribe()
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
// subjectMatches checks if a subject matches a pattern
func subjectMatches(pattern, subject string) bool {
	// Convert NATS wildcards to regex patterns
	if pattern == ">" {
		return true
	}
	
	// Split into tokens
	patternTokens := strings.Split(pattern, ".")
	subjectTokens := strings.Split(subject, ".")

	if len(patternTokens) > len(subjectTokens) {
		return false
	}

	for i, pt := range patternTokens {
		if pt == ">" {
			return true
		}
		if pt == "*" {
			continue
		}
		if i >= len(subjectTokens) || pt != subjectTokens[i] {
			return false
		}
	}

	return len(patternTokens) == len(subjectTokens)
}
