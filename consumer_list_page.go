package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
)

type ConsumerListPage struct {
	*tview.Flex
	Data           *ds.Data
	consumerList   *tview.List
	app            *tview.Application
	footerTxt      *tview.TextView
	streamName     string
	deleteConfirmConsumer string
	deleteConfirmTimer    *time.Timer
}

func NewConsumerListPage(app *tview.Application, data *ds.Data) *ConsumerListPage {
	cp := &ConsumerListPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	// Create header
	headerRow1 := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerRow2 := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerRow2.AddItem(createTextView("[ESC] Back [a] Add [e] Edit [i] Info [DEL] Delete [ESC] Back", tcell.ColorWhite), 0, 1, false)

	// Create consumer list
	cp.consumerList = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetMainTextColor(tcell.ColorWhite).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorWhite)
	cp.consumerList.SetBorder(true)
	cp.consumerList.SetBorderPadding(2, 0, 1, 1)
	cp.consumerList.SetTitle("Consumers")

	// Create footer
	cp.footerTxt = createTextView("", tcell.ColorWhite)

	// Add all components
	cp.AddItem(headerRow1, 1, 0, false).
		AddItem(headerRow2, 1, 0, false).
		AddItem(cp.consumerList, 0, 1, true).
		AddItem(cp.footerTxt, 1, 0, false)

	cp.setupInputCapture()

	return cp
}

func (cp *ConsumerListPage) setupInputCapture() {
	cp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			cp.goBack()
			return nil
		case tcell.KeyDelete:
			if cp.consumerList.GetItemCount() == 0 {
				cp.notify("No consumer selected", 3*time.Second, "error")
				return event
			}
			idx := cp.consumerList.GetCurrentItem()
			consumerName, _ := cp.consumerList.GetItemText(idx)
			
			if cp.deleteConfirmConsumer == consumerName {
				// Second press - execute delete
				cp.deleteConfirmTimer.Stop()
				cp.deleteConfirmConsumer = ""
				cp.notify("Delete consumer functionality coming soon...", 3*time.Second, "info")
			} else {
				// First press - start confirmation
				logger.Info("Delete consumer action triggered for: %s", consumerName)
				cp.startDeleteConfirmation(consumerName)
			}
			return nil
		default:
			switch event.Rune() {
			case 'a', 'A':
				logger.Info("Add consumer action triggered")
				pages.SwitchToPage("consumerAddPage")
				_, b := pages.GetFrontPage()
				addPage := b.(*ConsumerAddPage)
				addPage.redraw(&cp.Data.CurrCtx)
			case 'e', 'E':
				if cp.consumerList.GetItemCount() == 0 {
					cp.notify("No consumer selected", 3*time.Second, "error")
					return event
				}
				idx := cp.consumerList.GetCurrentItem()
				consumerName, _ := cp.consumerList.GetItemText(idx)
				logger.Info("Edit consumer action triggered for: %s", consumerName)
				cp.notify("Edit consumer functionality coming soon...", 3*time.Second, "info")
			case 'i', 'I':
				if cp.consumerList.GetItemCount() == 0 {
					cp.notify("No consumer selected", 3*time.Second, "error")
					return event
				}
				idx := cp.consumerList.GetCurrentItem()
				consumerName, _ := cp.consumerList.GetItemText(idx)
				logger.Info("Consumer info action triggered for: %s", consumerName)
				cp.notify("Consumer info functionality coming soon...", 3*time.Second, "info")
			}
		}
		return event
	})
}

func (cp *ConsumerListPage) startDeleteConfirmation(consumerName string) {
	cp.deleteConfirmConsumer = consumerName
	cp.notify("Press DEL again within 10 seconds to confirm deletion of '"+consumerName+"'", 10*time.Second, "warning")
	
	if cp.deleteConfirmTimer != nil {
		cp.deleteConfirmTimer.Stop()
	}
	
	cp.deleteConfirmTimer = time.NewTimer(10 * time.Second)
	go func() {
		<-cp.deleteConfirmTimer.C
		cp.deleteConfirmConsumer = ""
		cp.notify("Delete confirmation timed out", 3*time.Second, "info")
	}()
}

func (cp *ConsumerListPage) notify(message string, duration time.Duration, logLevel string) {
	cp.footerTxt.SetText(message)
	cp.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		cp.footerTxt.SetText("")
		cp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}

func (cp *ConsumerListPage) redraw(ctx *ds.Context) {
	cp.consumerList.Clear()

	// Get JetStream context
	js, err := ctx.Conn.JetStream()
	if err != nil {
		cp.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	consumersChan := js.Consumers(cp.streamName)
	if err != nil {
		cp.notify("Failed to get consumers: "+err.Error(), 3*time.Second, "error")
		return
	}

	for consumer := range consumersChan {
		cp.consumerList.AddItem(consumer.Name, "", 0, nil)
	}
	
}


func (cp *ConsumerListPage) goBack() {
	pages.SwitchToPage("streamListPage")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(&cp.Data.CurrCtx)
	cp.app.SetFocus(b) // Add this line
}