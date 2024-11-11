package main

import (
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nats-io/nats.go"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
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
	consumer      *nats.Subscription
	consumerMu    sync.Mutex
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
	// Header setup with simplified controls
	headerText := "[Esc] Back | [Tab] Next Field | [Alt+Enter] Send"
	headerRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		SetBorderPadding(1, 0, 1, 1)
	headerRow.AddItem(createTextView(headerText, tcell.ColorWhite), 0, 1, false)
	headerRow.SetTitle("Stream View")
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
	svp.logView.SetTitle(svp.Data.CurrCtx.LogFilePath)
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

	svp.SetBorderPadding(0, 0, 1, 1)
}

func (svp *StreamViewPage) setupInputCapture() {
	svp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			svp.goBack()
			return nil
		}
		return event
	})
}

func (svp *StreamViewPage) redraw(ctx *ds.Context) {
	svp.logView.Clear()
	svp.logView.SetTitle(ctx.LogFilePath)
	svp.createTemporaryConsumer()
	svp.app.SetFocus(svp.filterSubject)
}

func (svp *StreamViewPage) createTemporaryConsumer() {
	svp.consumerMu.Lock()
	defer svp.consumerMu.Unlock()

	js, err := svp.Data.CurrCtx.Conn.JetStream()
	if err != nil {
		svp.log("ERROR: Failed to get JetStream context: " + err.Error())
		return
	}

	// Clean up previous subscription if exists
	if svp.consumer != nil {
		if err := svp.consumer.Unsubscribe(); err != nil {
			svp.log("WARN: Error unsubscribing consumer: " + err.Error())
		}
		svp.consumer = nil
	}

	// Create ephemeral consumer subscription
	filterSubject := svp.filterSubject.GetText()
	if filterSubject == "" {
		filterSubject = ">" // Subscribe to all subjects if no filter
	}

	// Create new subscription that delivers new messages
	sub, err := js.Subscribe(filterSubject, func(msg *nats.Msg) {
		svp.displayMessage(msg)
		msg.Ack()
	}, nats.BindStream(svp.streamName),
	   nats.DeliverNew(),
	   nats.AckExplicit())
	if err != nil {
		svp.log("ERROR: Failed to create subscription: " + err.Error())
		return
	}

	svp.consumer = sub
	svp.log("INFO: Subscribed to: " + filterSubject + " (new messages only)")
}

func (svp *StreamViewPage) updateConsumerFilter() {
	if svp.consumer != nil {
		svp.consumer.Unsubscribe()
		svp.consumer = nil
	}
	svp.createTemporaryConsumer() // Recreate with new filter
}



func (svp *StreamViewPage) publishMessage() {
	js, err := svp.Data.CurrCtx.Conn.JetStream()
	if err != nil {
		svp.log("Failed to get JetStream context: "+err.Error())
		return
	}

	subject := svp.subjectName.GetText()
	if subject == "" {
		svp.log("ERROR: Subject cannot be empty")
		return
	}

	message := svp.txtArea.GetText()
	if message == "" {
		svp.log("ERROR: Message cannot be empty")
		return
	}

	svp.log("PUB[" + subject + "] " + message)
	svp.txtArea.SetText("", true)


	// Get stream info to check subjects
	stream, err := js.StreamInfo(svp.streamName)
	if err != nil {
		svp.log("ERROR: Failed to get stream info: " + err.Error())
		return
	}

	// Verify subject matches stream's subject filter
	subjectAllowed := false
	for _, s := range stream.Config.Subjects {
		if isValidSubject(subject) && subjectMatches(s, subject) {
			subjectAllowed = true
			break
		}
	}

	subjectsConfigStr := strings.Join(stream.Config.Subjects, ", ")

	if !subjectAllowed {
		svp.log("ERROR: Subject does not match stream's subject filter: " + subjectsConfigStr)
		return
	}

	_, err = js.Publish(subject, []byte(message))
	if err != nil {
		svp.log("ERROR: Failed to publish message: " + err.Error())
		return
	}

	// Lock for consumer operations
	svp.consumerMu.Lock()
	defer svp.consumerMu.Unlock()

	// Clean up existing consumer if it exists
	if svp.consumer != nil {
		if err := svp.consumer.Unsubscribe(); err != nil {
			svp.log("WARN: Error unsubscribing consumer: " + err.Error())
		}
		svp.consumer = nil
	}

	// Create new subscription starting from the published message
	filterSubject := svp.filterSubject.GetText()
	if filterSubject == "" {
		filterSubject = ">"
	}

	sub, err := js.PullSubscribe(filterSubject, "", 
		nats.BindStream(svp.streamName),
		nats.AckExplicit(),
		nats.DeliverLast())
	if err != nil {
		svp.log("ERROR: Failed to create subscription: " + err.Error())
		return
	}

	// Verify the subscription is valid
	if _, err := sub.ConsumerInfo(); err != nil {
		svp.log("ERROR: Failed to verify consumer: " + err.Error())
		sub.Unsubscribe()
		return
	}

	svp.log("INFO: Message published successfully")
}

func (svp *StreamViewPage) displayMessage(msg *nats.Msg) {
	timestamp := time.Now().Format("15:04:05.00000")
	text := timestamp + " [" + msg.Subject + "] " + string(msg.Data) + "\n"
	svp.logView.Write([]byte(text))
	svp.Data.CurrCtx.LogFile.Write([]byte(text))
	svp.logView.ScrollToEnd()
}

func (svp *StreamViewPage) log(message string) {
	hourMinSec := time.Now().Format("15:04:05.00000")
	logMessage := hourMinSec + " " + message + "\n"
	
	// Write to log view
	svp.logView.Write([]byte(logMessage))
	svp.logView.ScrollToEnd()
	
	// Write to log file
	svp.Data.CurrCtx.LogFile.WriteString(logMessage)
}

func (svp *StreamViewPage) goBack() {
	svp.consumerMu.Lock()
	if svp.consumer != nil {
		if err := svp.consumer.Unsubscribe(); err != nil {
			svp.log("WARN: Error unsubscribing consumer: " + err.Error())
		}
		svp.consumer = nil
	}
	svp.consumerMu.Unlock()
	
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
// isValidSubject checks if a subject string is valid according to NATS rules
func isValidSubject(subject string) bool {
	if subject == "" {
		return false
	}

	// Split into tokens
	tokens := strings.Split(subject, ".")
	
	for _, token := range tokens {
		if token == "" {
			return false // Empty token between dots
		}
		
		// Check for invalid characters
		for _, ch := range token {
			if !((ch >= 'a' && ch <= 'z') || 
				 (ch >= 'A' && ch <= 'Z') || 
				 (ch >= '0' && ch <= '9') || 
				 ch == '-' || ch == '_' || 
				 ch == '>' || ch == '*') {
				return false
			}
		}
	}
	return true
}

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
