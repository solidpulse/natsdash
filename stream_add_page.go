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
	headerRow := tview.NewFlex()
	headerRow.SetDirection(tview.FlexColumn)
	headerRow.SetBorderPadding(0, 0, 1, 1)

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
	footer := tview.NewFlex()
	footer.SetBorder(true)
	sap.footerTxt = createTextView("", tcell.ColorWhite)
	footer.AddItem(sap.footerTxt, 0, 1, false)
	sap.AddItem(footer, 3, 1, false)

	userFriendlyYAML := `# Stream Configuration

# Name of the stream (required)
name: my_stream

# Description of the stream (optional)
description: My Stream Description

# Subjects that messages can be published to (required)
# Examples: ["orders.*", "shipping.>", "customer.orders.*"]
subjects: 
  - my.subject.>

# Storage backend (required)
# Possible values: file, memory
storage: file

# Number of replicas for the stream
# Range: 1-5
num_replicas: 1

# Retention policy (required)
# Possible values: limits, interest, workqueue
retention: limits

# Discard policy when limits are reached
# Possible values: old, new
discard: old

# Maximum number of messages in the stream
# -1 for unlimited
max_msgs: -1

# Maximum number of bytes in the stream
# -1 for unlimited
max_bytes: -1

# Maximum age of messages
# Examples: 24h, 7d, 1y
max_age: 24h

# Maximum message size in bytes
# -1 for unlimited
max_msg_size: -1

# Maximum number of messages per subject
# -1 for unlimited
max_msgs_per_subject: -1

# Maximum number of consumers
# -1 for unlimited
max_consumers: -1
`
	sap.textArea.SetText(userFriendlyYAML, false)
}

func (sap *StreamAddPage) setupInputCapture() {
	sap.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			sap.notify("Loading......", 3*time.Second, "info")
			sap.goBack()
			return nil
		}
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			yamlText := sap.textArea.GetText()
			
			// Parse the YAML into our config struct
			var config StreamConfig
			err := yaml.Unmarshal([]byte(yamlText), &config)
			if err != nil {
				sap.notify("Invalid YAML configuration: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Connect to NATS
			conn, err := natsutil.Connect(&sap.Data.CurrCtx.CtxData)
			if err != nil {
				sap.notify("Failed to connect to NATS: "+err.Error(), 3*time.Second, "error")
				return nil
			}
			defer conn.Close()

			// Get JetStream context
			js, err := conn.JetStream()
			if err != nil {
				sap.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Convert to nats.StreamConfig
			streamConfig := nats.StreamConfig{
				Name:              config.Name,
				Description:       config.Description,
				Subjects:         config.Subjects,
				Retention:        config.Retention,
				MaxConsumers:     config.MaxConsumers,
				MaxMsgs:          config.MaxMsgs,
				MaxBytes:         config.MaxBytes,
				Discard:          config.Discard,
				MaxAge:           config.MaxAge,
				MaxMsgsPerSubject: config.MaxMsgsPerSubject,
				MaxMsgSize:       config.MaxMsgSize,
				Storage:          config.Storage,
				Replicas:         config.Replicas,
			}

			// Create the stream
			_, err = js.AddStream(&streamConfig)
			if err != nil {
				sap.notify("Failed to create stream: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			sap.notify("Stream created successfully", 3*time.Second, "info")
			sap.goBack()
			return nil
		}
		return event
	})
}

func (sap *StreamAddPage) goBack() {
	pages.SwitchToPage("streamListPage")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(&sap.Data.CurrCtx)
	sap.app.SetFocus(b)
}

func (sap *StreamAddPage) redraw(ctx *ds.Context) {

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
