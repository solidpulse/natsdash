package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createColoredTextView(text string, color tcell.Color) *tview.TextView {
	return tview.NewTextView().
		SetTextColor(color).
		SetText(text).
		SetDynamicColors(true)
}

func createTextView(text string, color tcell.Color) *tview.TextView {
	return tview.NewTextView().
		SetTextColor(color).
		SetText(text)
}
