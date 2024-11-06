package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nats-io/nats.go"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
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

		// Convert to JSON
		jsonBytes, err := json.MarshalIndent(consumer.Config, "", "    ")
		if err != nil {
			cap.notify("Failed to convert config to JSON: "+err.Error(), 3*time.Second, "error")
			return
		}

		cap.txtArea.SetText(string(jsonBytes), true)
	} else {
		// Set default template for new consumer
		defaultConfig := fmt.Sprintf(`{
    "deliver_policy": "all",
    "ack_policy": "explicit",
    "replay_policy": "instant",
    "max_deliver": 1
}`)
		cap.txtArea.SetText(defaultConfig, true)
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

	// Parse the configuration
	var config nats.ConsumerConfig
	if err := json.Unmarshal([]byte(cap.txtArea.GetText()), &config); err != nil {
		cap.notify("Invalid configuration: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Set the name from the previous screen
	config.Name = cap.consumerName

	// Create or update the consumer
	var consumer *nats.ConsumerInfo
	if cap.isEdit {
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
	pages.SwitchToPage("consumerListPage")
	_, p := pages.GetFrontPage()
	listPage := p.(*ConsumerListPage)
	listPage.redraw(&cap.Data.CurrCtx)
}
