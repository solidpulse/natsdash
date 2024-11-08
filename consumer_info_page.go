package main

import (
	"encoding/json"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
	"gopkg.in/yaml.v2"
)

type ConsumerInfoPage struct {
	*tview.Flex
	Data         *ds.Data
	app          *tview.Application
	txtArea      *tview.TextArea
	footerTxt    *tview.TextView
	streamName   string
	consumerName string
}

func NewConsumerInfoPage(app *tview.Application, data *ds.Data) *ConsumerInfoPage {
	cip := &ConsumerInfoPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	// Create header
	headerRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerRow.AddItem(createTextView("[ESC] Back", tcell.ColorWhite), 0, 1, false)

	// Create text area
	cip.txtArea = tview.NewTextArea()
	cip.txtArea.SetBorder(true)
	
	// Create footer
	cip.footerTxt = createTextView("", tcell.ColorWhite)

	// Add all components
	cip.AddItem(headerRow, 1, 0, false).
		AddItem(cip.txtArea, 0, 1, true).
		AddItem(cip.footerTxt, 1, 0, false)

	cip.setupInputCapture()

	return cip
}

func (cip *ConsumerInfoPage) redraw(ctx *ds.Context) {
	cip.txtArea.SetTitle("Consumer Info: " + cip.consumerName)

	// Get JetStream context
	js, err := ctx.Conn.JetStream()
	if err != nil {
		cip.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Get consumer info
	consumer, err := js.ConsumerInfo(cip.streamName, cip.consumerName)
	if err != nil {
		cip.notify("Failed to get consumer info: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Convert to map first
	var infoMap map[string]interface{}
	jsonBytes, err := json.Marshal(consumer)
	if err != nil {
		cip.notify("Failed to convert info: "+err.Error(), 3*time.Second, "error")
		return
	}
	if err := json.Unmarshal(jsonBytes, &infoMap); err != nil {
		cip.notify("Failed to process info: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Convert duration fields to strings
	if config, ok := infoMap["config"].(map[string]interface{}); ok {
		if ackWait, ok := config["ack_wait"].(float64); ok {
			config["ack_wait"] = time.Duration(ackWait).String()
		}
		if idleHeartbeat, ok := config["idle_heartbeat"].(float64); ok {
			config["idle_heartbeat"] = time.Duration(idleHeartbeat).String()
		}
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(infoMap)
	if err != nil {
		cip.notify("Failed to convert info to YAML: "+err.Error(), 3*time.Second, "error")
		return
	}

	cip.txtArea.SetText(string(yamlBytes), false)
}

func (cip *ConsumerInfoPage) setupInputCapture() {
	cip.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cip.goBack()
			return nil
		}
		return event
	})
}

func (cip *ConsumerInfoPage) goBack() {
	pages.SwitchToPage("consumerListPage")
	_, b := pages.GetFrontPage()
	b.(*ConsumerListPage).streamName = cip.streamName
	b.(*ConsumerListPage).redraw(&cip.Data.CurrCtx)
	cip.app.SetFocus(b)
}

func (cip *ConsumerInfoPage) notify(message string, duration time.Duration, logLevel string) {
	cip.footerTxt.SetText(message)
	cip.footerTxt.SetTextColor(getLogLevelColor(logLevel))
	logger.Info(message)

	go func() {
		time.Sleep(duration)
		cip.footerTxt.SetText("")
		cip.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}
