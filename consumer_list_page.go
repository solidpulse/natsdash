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
	headerRow2 := tview.NewFlex().SetDirection(tview.FlexRow)
	txtViewHeader := createTextView("[ESC] Back [a] Add [e] Edit [i] Info [d] Delete", tcell.ColorWhite)
	txtViewHeader.SetBorderPadding(1,1,1,1)
	headerRow2.AddItem(txtViewHeader, 0, 1, false)

	// Create consumer list
	cp.consumerList = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetMainTextColor(tcell.ColorWhite).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorWhite)
	cp.consumerList.SetBorder(true)
	cp.consumerList.SetBorderPadding(0, 0, 1, 1)
	cp.consumerList.SetTitle("Consumers")

	// Create footer
	cp.footerTxt = createTextView("", tcell.ColorWhite)
	cp.footerTxt.SetBorder(true)

	// Add all components
	cp.
		AddItem(headerRow2, 3, 0, false).
		AddItem(cp.consumerList, 0, 1, true).
		AddItem(cp.footerTxt, 3, 0, false)

	cp.setupInputCapture()

	return cp
}

func (cp *ConsumerListPage) setupInputCapture() {
	cp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			cp.goBack()
			return nil
		default:
			switch event.Rune() {
			case 'a', 'A':
				logger.Info("Add consumer action triggered")
				pages.SwitchToPage("consumerAddPage")
				_, b := pages.GetFrontPage()
				addPage := b.(*ConsumerAddPage)
				addPage.streamName = cp.streamName
				addPage.isEdit = false	
				addPage.redraw(&cp.Data.CurrCtx)
			case 'e', 'E':
				if cp.consumerList.GetItemCount() == 0 {
					cp.notify("No consumer selected", 3*time.Second, "error")
					return event
				}
				idx := cp.consumerList.GetCurrentItem()
				consumerName, _ := cp.consumerList.GetItemText(idx)
				logger.Info("Add consumer action triggered")
				pages.SwitchToPage("consumerAddPage")
				_, b := pages.GetFrontPage()
				editPage := b.(*ConsumerAddPage)
				editPage.streamName = cp.streamName
				editPage.isEdit = true
				editPage.consumerName = consumerName
				editPage.redraw(&cp.Data.CurrCtx)
			case 'd', 'D':
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
					cp.deleteConsumer(consumerName)
				} else {
					// First press - start confirmation
					logger.Info("Delete consumer action triggered for: %s", consumerName)
					cp.startDeleteConfirmation(consumerName)
				}
				return nil
			case 'i', 'I':
				if cp.consumerList.GetItemCount() == 0 {
					cp.notify("No consumer selected", 3*time.Second, "error")
					return event
				}
				idx := cp.consumerList.GetCurrentItem()
				consumerName, _ := cp.consumerList.GetItemText(idx)
				logger.Info("Consumer info action triggered for: %s", consumerName)
				pages.SwitchToPage("consumerInfoPage")
				_, b := pages.GetFrontPage()
				infoPage := b.(*ConsumerInfoPage)
				infoPage.streamName = cp.streamName
				infoPage.consumerName = consumerName
				infoPage.redraw(&cp.Data.CurrCtx)
			}
		}
		return event
	})
}

func (cp *ConsumerListPage) startDeleteConfirmation(consumerName string) {
	cp.deleteConfirmConsumer = consumerName
	cp.notify("Press 'd' again within 10 seconds to confirm deletion of '"+consumerName+"'", 10*time.Second, "warning")
	
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

	for consumer := range consumersChan {
		cp.consumerList.AddItem(consumer.Name, "", 0, nil)
	}
	
}


func (cp *ConsumerListPage) deleteConsumer(consumerName string) {
	// Get JetStream context
	js, err := cp.Data.CurrCtx.Conn.JetStream()
	if err != nil {
		cp.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Delete the consumer
	err = js.DeleteConsumer(cp.streamName, consumerName)
	if err != nil {
		cp.notify("Failed to delete consumer: "+err.Error(), 3*time.Second, "error")
		return
	}

	cp.notify("Consumer '"+consumerName+"' deleted successfully", 3*time.Second, "info")
	cp.redraw(&cp.Data.CurrCtx)
}

func (cp *ConsumerListPage) goBack() {
	pages.SwitchToPage("streamListPage")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(&cp.Data.CurrCtx)
	cp.app.SetFocus(b) // Add this line
}
