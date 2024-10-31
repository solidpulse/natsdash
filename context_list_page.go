package main

import (
	"github.com/evnix/natsdash/ds"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContextPage struct {
	*tview.Flex
	Data        *ds.Data
	ctxListView *tview.List
}

func NewContextPage(data *ds.Data) *ContextPage {
	cp := &ContextPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	cp.Data = data

	cp.setupUI()
	cp.setupInputCapture()

	return cp
}

func (cp *ContextPage) setupUI() {
	// Header setup
	headerRow := createContexListHeaderRow()
	cp.AddItem(headerRow, 0, 4, false)

	// Context list setup
	ctxListBox := tview.NewFlex()
	ctxListBox.SetTitle("Contexts").SetBorder(true)
	cp.ctxListView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	ctxListBox.AddItem(cp.ctxListView, 0, 20, false)
	cp.AddItem(ctxListBox, 0, 18, false)

	// Footer setup
	footer := tview.NewFlex().SetBorder(true)
	cp.AddItem(footer, 0, 1, false)

	cp.SetBorderPadding(1, 1, 1, 1)
}

func (cp *ContextPage) setupInputCapture() {
	cp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			app.Stop()
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			app.SetFocus(cp.ctxListView)
		case tcell.KeyDelete:
			idx := cp.ctxListView.GetCurrentItem()
			//remove from contexts
			cp.Data.Contexts = append(cp.Data.Contexts[:idx], cp.Data.Contexts[idx+1:]...)
			//save to file
			cp.Data.SaveToFile()
			//redraw
			cp.Redraw()
		}
		if event.Rune() == 'a' || event.Rune() == 'A' {
			data.CurrCtx = ds.Context{}
			pages.SwitchToPage("contextFormPage")
			_, b := pages.GetFrontPage()
			b.(*ContextFormPage).redraw(&data.CurrCtx)
		} else if event.Rune() == 'e' || event.Rune() == 'E' {
			idx := cp.ctxListView.GetCurrentItem()
			data.CurrCtx = cp.Data.Contexts[idx]
			pages.SwitchToPage("contextFormPage")
			_, b := pages.GetFrontPage()
			b.(*ContextFormPage).redraw(&data.CurrCtx)
		} else if event.Rune() == 'i' || event.Rune() == 'I' {
			idx := cp.ctxListView.GetCurrentItem()
			data.CurrCtx = cp.Data.Contexts[idx]
			pages.SwitchToPage("serverInfoPage")
			_, b := pages.GetFrontPage()
			b.(*ServerInfoPage).redraw(&data.CurrCtx)
		} else if event.Rune() == 'n' || event.Rune() == 'N' {
			idx := cp.ctxListView.GetCurrentItem()
			data.CurrCtx = cp.Data.Contexts[idx]
			pages.SwitchToPage("natsPage")
			_, b := pages.GetFrontPage()
			b.(*NatsPage).redraw(&data.CurrCtx)
		}
		return event
	})
}

func (cp *ContextPage) Redraw() {
	cp.ctxListView.Clear()
	for _, ctx := range cp.Data.Contexts {
		cp.ctxListView.AddItem(ctx.Name, "", 0, nil)
	}
}

func createContexListHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createColoredTextView("[white:green] NATS [yellow:white] DASH ", tcell.ColorWhite), 0, 1, false)
	headerRow1.AddItem(createTextView("[a] Add", tcell.ColorWhite), 0, 1, false)
	headerRow1.AddItem(createTextView("[e] Edit", tcell.ColorWhite), 0, 1, false)

	headerRow2 := tview.NewFlex()
	headerRow2.SetDirection(tview.FlexRow)
	headerRow2.SetBorder(false)

	headerRow2.AddItem(createTextView("[i] Info", tcell.ColorWhite), 0, 1, false)
	headerRow2.AddItem(createTextView("[n] Core NATS", tcell.ColorWhite), 0, 1, false)
	headerRow2.AddItem(createTextView("[j] Jetstream", tcell.ColorWhite), 0, 1, false)
	headerRow2.AddItem(createTextView("[Del] Delete", tcell.ColorWhite), 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.AddItem(headerRow2, 0, 1, false)
	headerRow.SetTitle("NATS-DASH")

	return headerRow
}
