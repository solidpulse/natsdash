package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContextFormPage struct {
	*tview.Flex
	form *tview.Form
}

func NewContextFormPage() *ContextFormPage {
	cfp := &ContextFormPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	cfp.setupUI()
	cfp.setupInputCapture()

	return cfp
}

func (cfp *ContextFormPage) setupUI() {
	// Header setup
	headerRow := createHeaderRow()
	cfp.AddItem(headerRow, 0, 4, false)

	// Form setup
	cfp.form = tview.NewForm()
	cfp.form.SetTitle("Add Context").
		SetBorder(true)
	cfp.form.
		AddInputField("Name", "", 0, nil, nil).
		AddInputField("URL", "nats://", 0, nil, nil).
		AddButton("Save", cfp.saveContext).
		AddButton("Cancel", cfp.cancelForm)
	cfp.AddItem(cfp.form, 0, 18, true)

	// Footer setup
	footer := tview.NewFlex().SetBorder(true)
	cfp.AddItem(footer, 0, 1, false)

	cfp.SetBorderPadding(1, 1, 1, 1)
}

func (cfp *ContextFormPage) setupInputCapture() {
	cfp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cfp.goBackToContextPage()
			return nil
		}
		return event
	})
}

func (cfp *ContextFormPage) saveContext() {
	name := cfp.form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
	url := cfp.form.GetFormItemByLabel("URL").(*tview.InputField).GetText()
	// TODO: Implement save functionality
	fmt.Printf("Saving context: Name=%s, URL=%s\n", name, url)
	cfp.goBackToContextPage()
}

func (cfp *ContextFormPage) cancelForm() {
	cfp.form.GetFormItemByLabel("Name").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("URL").(*tview.InputField).SetText("")
	cfp.goBackToContextPage()
}

func (cfp *ContextFormPage) goBackToContextPage() {
	pages.SwitchToPage("contexts")
}
