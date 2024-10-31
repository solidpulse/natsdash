package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContextPage struct {
	*tview.Flex
	ctxListView *tview.List
}

func NewContextPage() *ContextPage {
	cp := &ContextPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	cp.setupUI()
	cp.setupInputCapture()

	return cp
}

func (cp *ContextPage) setupUI() {
	// Header setup
	headerRow := createHeaderRow()
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
		}
		if event.Rune() == 'a' || event.Rune() == 'A' {
			pages.SwitchToPage("contextFormPage")
		}
		return event
	})
}

func (cp *ContextPage) UpdateData(contexts []string) {
	cp.ctxListView.Clear()
	for _, ctx := range contexts {
		cp.ctxListView.AddItem(ctx, "", 0, nil)
	}
}
