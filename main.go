package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()

func main() {
	fmt.Println("Hello, world.")

	var pages = tview.NewPages()
	pages.AddPage("contexts", GetContextPage(pages), true, true)
	pages.AddPage("contextFormPage", GetContextFormPage(pages), true, false)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}

func GetContextPage(pages *tview.Pages) *tview.Flex {
	var contextPageFlex = tview.NewFlex()
	contextPageFlex.SetDirection(tview.FlexRow)

	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.SetDirection(tview.FlexColumn)
	headerRow.SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	var text = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetText("[white:green] NATS [yellow:white] DASH ").SetDynamicColors(true)

	var addText = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetText("[a] Add")
	addText.SetBorderPadding(0, 0, 0, 0)

	var editText = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetText("[e] Edit")

	var deleteText = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetText("[Ctrl+Del] Delete")

	headerRow2 := tview.NewFlex()
	headerRow2.SetDirection(tview.FlexRow)
	headerRow2.SetBorder(false)
	// var blankTxt = tview.NewTextView().
	// 	SetTextColor(tcell.ColorWhite).
	// 	SetText("        ")
	var natsText = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetText("[n] Core NATS")

	var jetTxt = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetText("[j] Jetstream")

	headerRow2.AddItem(natsText, 0, 1, false)
	headerRow2.AddItem(jetTxt, 0, 1, false)
	// headerRow2.AddItem(blankTxt, 0, 1, false)

	headerRow1.AddItem(text, 0, 1, false)
	headerRow1.AddItem(addText, 0, 1, false)
	headerRow1.AddItem(editText, 0, 1, false)
	headerRow2.AddItem(deleteText, 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.AddItem(headerRow2, 0, 1, false)
	// headerRow.AddItem(jetTxt, 0, 1, false)
	headerRow.SetTitle("NATS-DASH")
	contextPageFlex.AddItem(headerRow, 0, 4, false)

	ctxListBox := tview.NewFlex()
	ctxListBox.SetTitle("Contexts")
	ctxListBox.SetBorder(true)
	ctxListView := tview.NewList()
	ctxListView.ShowSecondaryText(false)
	ctxListView.SetHighlightFullLine(true)
	ctxListView.AddItem("ctx1", "", 0, nil)
	ctxListView.AddItem("ctx2", "", 0, nil)

	ctxListBox.AddItem(ctxListView, 0, 20, false)

	contextPageFlex.AddItem(ctxListBox, 0, 18, false)
	contextPageFlex.SetBorderPadding(1, 1, 1, 1)
	footer := tview.NewFlex()
	footer.SetBorder(true)

	contextPageFlex.AddItem(footer, 0, 1, false)
	contextPageFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.Stop()
			return nil
		} else if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			app.SetFocus(ctxListView)
		} else if event.Rune() == 'a' || event.Rune() == 'A' {
			pages.AddAndSwitchToPage("contextFormPage", GetContextFormPage(pages), true)
		}
		return event
	})

	return contextPageFlex
}

func GetContextFormPage(pages *tview.Pages) *tview.Flex {
	var contextPageFlex = tview.NewFlex()
	contextPageFlex.SetDirection(tview.FlexRow)

	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.SetDirection(tview.FlexColumn)
	headerRow.SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	var addText = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetText("[Esc] Back")
	addText.SetBorderPadding(0, 0, 0, 0)

	headerRow1.AddItem(addText, 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.SetTitle("NATS-DASH")
	contextPageFlex.AddItem(headerRow, 0, 4, false)

	form := tview.NewForm()
	form.SetTitle("Add Context")
	form.SetBorder(true)

	form.AddInputField("Name", "", 0, nil, nil)
	form.AddInputField("URL", "nats://", 0, nil, nil)

	form.AddButton("Save", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		url := form.GetFormItem(1).(*tview.InputField).GetText()
		fmt.Println(name, url)
		// TODO: Implement save functionality
	})

	form.AddButton("Cancel", func() {
		form.GetFormItem(0).(*tview.InputField).SetText("")
		form.GetFormItem(1).(*tview.InputField).SetText("")
	})

	contextPageFlex.AddItem(form, 0, 18, true)
	contextPageFlex.SetBorderPadding(1, 1, 1, 1)

	footer := tview.NewFlex()
	footer.SetBorder(true)

	contextPageFlex.AddItem(footer, 0, 1, false)

	contextPageFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("contextFormPage")
			pages.ShowPage("contexts")
			return nil
		}
		return event
	})

	return contextPageFlex
}
