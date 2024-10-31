package main

import (
	"github.com/evnix/natsdash/ds"
	"github.com/evnix/natsdash/natsutil"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
)

type ContextFormPage struct {
	*tview.Flex
	Data     *ds.Data
	form     *tview.Form
	currUUID string
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
	headerRow := createContextFormHeaderRow()
	cfp.AddItem(headerRow, 0, 4, false)

	// Form setup
	cfp.form = tview.NewForm()
	cfp.form.SetTitle("Add Context").
		SetBorder(true)
	cfp.form.
		AddInputField("Name", "", 0, nil, nil).
		AddInputField("URL", "nats://", 0, nil, nil).
		AddTextView("", "", 0, 1, false, false).
		AddButton("Save", cfp.saveContext).
		AddButton("Cancel", cfp.cancelForm)
	cfp.AddItem(cfp.form, 0, 18, true)

	// Footer setup
	footer := tview.NewFlex().SetBorder(true)
	cfp.AddItem(footer, 0, 1, false)

	cfp.SetBorderPadding(1, 1, 1, 1)
}

func (cfp *ContextFormPage) redraw(ctx *ds.Context) {
	cfp.currUUID = ctx.UUID
	if cfp.currUUID != "" {
		cfp.form.GetFormItemByLabel("Name").(*tview.InputField).SetText(ctx.Name)
		cfp.form.GetFormItemByLabel("URL").(*tview.InputField).SetText(ctx.URL)
	} else {
		cfp.form.GetFormItemByLabel("Name").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("URL").(*tview.InputField).SetText("nats://")
	}
	errTxt := cfp.form.GetFormItem(2).(*tview.TextView)
	errTxt.SetText("")
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
	errTxt := cfp.form.GetFormItem(2).(*tview.TextView)
	errTxt.SetText("Connecting to server...")

	uuid := uuid.New().String()
	newCtx := ds.Context{UUID: uuid, Name: name, URL: url}

	go func() {
		err := natsutil.TestConnect(url)
		if err != nil {
			errTxt.SetText(err.Error())
			return
		}

		if cfp.currUUID != "" {
			for i := range cfp.Data.Contexts {
				if cfp.Data.Contexts[i].UUID == cfp.currUUID {
					cfp.Data.Contexts[i].Name = name
					cfp.Data.Contexts[i].URL = url
					cfp.Data.CurrCtx = cfp.Data.Contexts[i]
					break
				}
			}
		} else {
			cfp.Data.Contexts = append(cfp.Data.Contexts, newCtx)
			cfp.Data.CurrCtx = newCtx
		}

		cfp.Data.SaveToFile()
		cfp.goBackToContextPage()
	}()

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
	cfp.app.SetFocus(b) // Add this line
}

func createContextFormHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createTextView("[Esc] Back", tcell.ColorWhite), 0, 1, false)

	headerRow2 := tview.NewFlex()
	headerRow2.SetDirection(tview.FlexRow)
	headerRow2.SetBorder(false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.AddItem(headerRow2, 0, 1, false)
	headerRow.SetTitle("NATS-DASH")

	return headerRow
}
