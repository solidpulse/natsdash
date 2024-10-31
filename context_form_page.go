package main

import (
	"github.com/evnix/natsdash/ds"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
)

type ContextFormPage struct {
	*tview.Flex
	Data *ds.Data
	form *tview.Form
}

func NewContextFormPage(data *ds.Data) *ContextFormPage {
	cfp := &ContextFormPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	cfp.Data = data
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

	uuid := uuid.New().String()
	newCtx := ds.Context{UUID: uuid, Name: name, URL: url}
	cfp.Data.Contexts = append(cfp.Data.Contexts, newCtx)
	cfp.Data.SaveToFile()
	cfp.goBackToContextPage()
}

func (cfp *ContextFormPage) cancelForm() {
	cfp.form.GetFormItemByLabel("Name").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("URL").(*tview.InputField).SetText("")
	cfp.goBackToContextPage()
}

func (cfp *ContextFormPage) goBackToContextPage() {
	pages.SwitchToPage("contexts")
	_, b := pages.GetFrontPage()
	b.(*ContextPage).Redraw()
}
