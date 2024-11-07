package main

import (
	"encoding/json"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nats-io/nats.go"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
	"gopkg.in/yaml.v2"
)

type ConsumerAddPage struct {
	*tview.Flex
	Data         *ds.Data
	app          *tview.Application
	txtArea      *tview.TextArea
	footerTxt    *tview.TextView
	streamName   string
	consumerName string
	isEdit       bool
}

func NewConsumerAddPage(app *tview.Application, data *ds.Data) *ConsumerAddPage {
	cap := &ConsumerAddPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	// Create header
	headerRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerRow.AddItem(createTextView("[ESC] Back [Alt+Enter] Save", tcell.ColorWhite), 0, 1, false)

	// Create text area
	cap.txtArea = tview.NewTextArea()
	cap.txtArea.SetBorder(true)
	
	// Create footer
	cap.footerTxt = createTextView("", tcell.ColorWhite)

	// Add all components
	cap.AddItem(headerRow, 1, 0, false).
		AddItem(cap.txtArea, 0, 1, true).
		AddItem(cap.footerTxt, 1, 0, false)

	cap.setupInputCapture()

	return cap
}

func (cap *ConsumerAddPage) redraw(ctx *ds.Context) {
	// Update the title based on mode
	title := "Add Consumer"
	if cap.isEdit {
		title = "Edit Consumer: " + cap.consumerName
	}
	cap.txtArea.SetTitle(title)

	if cap.isEdit {
		// Get existing consumer config
		js, err := ctx.Conn.JetStream()
		if err != nil {
			cap.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
			return
		}

		consumer, err := js.ConsumerInfo(cap.streamName, cap.consumerName)
		if err != nil {
			cap.notify("Failed to get consumer info: "+err.Error(), 3*time.Second, "error")
			return
		}

		// Convert to YAML
		yamlBytes, err := yaml.Marshal(consumer.Config)
		if err != nil {
			cap.notify("Failed to convert config to JSON: "+err.Error(), 3*time.Second, "error")
			return
		}

		cap.txtArea.SetText(string(yamlBytes), true)
	} else {
		// Set default template for new consumer
		defaultConfig := `# Name of the consumer (required)
# Will be set automatically from previous screen
name: ""

# Durable name for the consumer (optional)
# Makes this a durable consumer that survives restarts
durable_name: "NEW"

# Pull mode configuration (optional)
# true: pull-based / false: push-based
pull: true

# Subject filter for the consumer (required)
# Examples: "ORDERS.*", "ORDERS.>", "ORDERS.*.received"
filter_subject: "ORDERS.received"

# Delivery policy configuration (optional)
# Only one of these should be true
deliver_all: true        # Deliver all available messages
deliver_last: false      # Deliver only the last message
# Other options (set all false if using deliver_all):
# deliver_new: false   # Only new messages
# deliver_by_start_sequence: false  # Start from specific sequence
# deliver_by_start_time: false      # Start from specific time

# Acknowledgment policy (required)
# Options: "none", "all", "explicit"
ack_policy: "explicit"

# Acknowledgment wait time (optional)
# How long to wait for ack before redelivery
# Format: "30s", "1m", "1h"
ack_wait: "30s"

# Replay policy (required)
# Options: "instant", "original"
replay_policy: "instant"

# Maximum delivery attempts (optional)
# How many times to attempt delivery before giving up
max_deliver: 20

# Sampling rate percentage (optional)
# 1-100, where 100 means all messages
sample_freq: 100

# Other available options:
# max_ack_pending: 1000    # Maximum pending acks
# max_waiting: 512         # Maximum waiting pulls
# max_batch: 100          # Maximum batch size for pull
# idle_heartbeat: "30s"   # Idle heartbeat interval
# flow_control: false     # Enable flow control
# headers_only: false     # Deliver only headers`
		cap.txtArea.SetText(defaultConfig, false)
	}
}
	
func (cp *ConsumerAddPage) goBack() {
	pages.SwitchToPage("consumerListPage")
	_, b := pages.GetFrontPage()
	b.(*ConsumerListPage).streamName = cp.streamName
	b.(*ConsumerListPage).redraw(&cp.Data.CurrCtx)
	cp.app.SetFocus(b) // Add this line
}

func (cp *ConsumerAddPage) notify(message string, duration time.Duration, logLevel string) {
	cp.footerTxt.SetText(message)
	cp.footerTxt.SetTextColor(getLogLevelColor(logLevel))
	logger.Info(message)

	go func() {
		time.Sleep(duration)
		cp.footerTxt.SetText("")
		cp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}
func (cap *ConsumerAddPage) setupInputCapture() {
	cap.txtArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("consumerListPage")
			return nil
		}
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			cap.saveConsumer()
			return nil
		}
		return event
	})
}
func (cap *ConsumerAddPage) saveConsumer() {
	// Get JetStream context
	js, err := cap.Data.CurrCtx.Conn.JetStream()
	if err != nil {
		cap.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	// First convert YAML to map
	var yamlData map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(cap.txtArea.GetText()), &yamlData); err != nil {
		cap.notify("Invalid YAML configuration: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Convert YAML map to JSON-compatible map
	jsonMap := convertToStringMap(yamlData)

	// Convert to JSON
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		cap.notify("Failed to process configuration: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Parse JSON into consumer config to use NATS struct tags
	var config nats.ConsumerConfig
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		cap.notify("Invalid configuration: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Create or update the consumer
	var consumer *nats.ConsumerInfo
	if cap.isEdit {
		config.Name = cap.consumerName
		consumer, err = js.UpdateConsumer(cap.streamName, &config)
	} else {
		consumer, err = js.AddConsumer(cap.streamName, &config)
	}

	if err != nil {
		cap.notify("Failed to save consumer: "+err.Error(), 3*time.Second, "error")
		return
	}

	cap.notify("Consumer "+consumer.Name+" saved successfully", 3*time.Second, "info")
	
	// Switch back to consumer list
	cap.goBack()
}

// Helper function to convert YAML map to JSON-compatible map
func convertToStringMap(m map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range m {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[k.(string)] = convertToStringMap(v2)
		case []interface{}:
			res[k.(string)] = convertToStringSlice(v2)
		default:
			res[k.(string)] = v
		}
	}
	return res
}

// Helper function to convert YAML slice to JSON-compatible slice
func convertToStringSlice(s []interface{}) []interface{} {
	res := make([]interface{}, len(s))
	for i, v := range s {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[i] = convertToStringMap(v2)
		case []interface{}:
			res[i] = convertToStringSlice(v2)
		default:
			res[i] = v
		}
	}
	return res
}
