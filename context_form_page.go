package main

import (
	"time"

	"github.com/evnix/natsdash/ds"
	"github.com/evnix/natsdash/natsutil"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContextFormPage struct {
	*tview.Flex
	Data     *ds.Data
	form     *tview.Form
	currName string
	app      *tview.Application // Add this line
	footerTxt      *tview.TextView
}

func NewContextFormPage(app *tview.Application, data *ds.Data) *ContextFormPage {
	cfp := &ContextFormPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app, // Add this line
	}

	cfp.Data = data
	cfp.setupUI()
	cfp.setupInputCapture()

	// Establish NATS connection when the context_form page opens
	go func() {
		if data.CurrCtx.CtxData.URL != "" {
			conn, err := natsutil.Connect(&data.CurrCtx.CtxData)
			if err != nil {
				// Handle error
				return
			}
			data.CurrCtx.Conn = conn
		}
	}()

	return cfp
}

func (cfp *ContextFormPage) setupUI() {
	// Header setup
	headerRow := createContextFormHeaderRow()
	cfp.AddItem(headerRow, 0, 4, false)

	// Form setup
	cfp.form = createContextForm(&cfp.Data.CurrCtx)
	cfp.form.AddButton("Save", cfp.saveContext).
		AddButton("Cancel", cfp.cancelForm)
	cfp.AddItem(cfp.form, 0, 18, true)

	// Footer setup
	footer := tview.NewFlex()
	footer.SetBorder(true)
	cfp.footerTxt = createTextView("", tcell.ColorWhite)
	footer.AddItem(cfp.footerTxt, 0, 1, false)
	cfp.AddItem(footer, 0, 1, false)

	cfp.SetBorderPadding(1, 1, 1, 1)
}

func (cfp *ContextFormPage) redraw(ctx *ds.Context) {
	cfp.currName = ctx.Name
	if cfp.currName != "" {
		cfp.form.GetFormItemByLabel("Name").(*tview.InputField).SetText(ctx.Name)
		cfp.form.GetFormItemByLabel("Description").(*tview.InputField).SetText(ctx.CtxData.Description)
		cfp.form.GetFormItemByLabel("URL").(*tview.InputField).SetText(ctx.CtxData.URL)
		cfp.form.GetFormItemByLabel("Token").(*tview.InputField).SetText(ctx.CtxData.Token)
		cfp.form.GetFormItemByLabel("User").(*tview.InputField).SetText(ctx.CtxData.User)
		cfp.form.GetFormItemByLabel("Password").(*tview.InputField).SetText(ctx.CtxData.Password)
		cfp.form.GetFormItemByLabel("Creds").(*tview.InputField).SetText(ctx.CtxData.Creds)
		cfp.form.GetFormItemByLabel("Nkey").(*tview.InputField).SetText(ctx.CtxData.Nkey)
		cfp.form.GetFormItemByLabel("Cert").(*tview.InputField).SetText(ctx.CtxData.Cert)
		cfp.form.GetFormItemByLabel("Key").(*tview.InputField).SetText(ctx.CtxData.Key)
		cfp.form.GetFormItemByLabel("CA").(*tview.InputField).SetText(ctx.CtxData.CA)
		cfp.form.GetFormItemByLabel("NSC").(*tview.InputField).SetText(ctx.CtxData.NSC)
		cfp.form.GetFormItemByLabel("Jetstream Domain").(*tview.InputField).SetText(ctx.CtxData.JetstreamDomain)
		cfp.form.GetFormItemByLabel("Jetstream API Prefix").(*tview.InputField).SetText(ctx.CtxData.JetstreamAPIPrefix)
		cfp.form.GetFormItemByLabel("Jetstream Event Prefix").(*tview.InputField).SetText(ctx.CtxData.JetstreamEventPrefix)
		cfp.form.GetFormItemByLabel("Inbox Prefix").(*tview.InputField).SetText(ctx.CtxData.InboxPrefix)
	} else {
		cfp.form.GetFormItemByLabel("Name").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Description").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("URL").(*tview.InputField).SetText("nats://")
		cfp.form.GetFormItemByLabel("Token").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("User").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Password").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Creds").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Nkey").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Cert").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Key").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("CA").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("NSC").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Jetstream Domain").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Jetstream API Prefix").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Jetstream Event Prefix").(*tview.InputField).SetText("")
		cfp.form.GetFormItemByLabel("Inbox Prefix").(*tview.InputField).SetText("")
	}
	cfp.notify("", 1*time.Second, "info")
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
	cfp.notify("Connecting to server...", 5*time.Second, "info")

	newCtx := ds.Context{ Name: name, CtxData: ds.NatsCliContext{URL: url}}

	go func() {
		err := natsutil.TestConnect(url)
		if err != nil {
			cfp.notify(err.Error(), 5*time.Second, "error")
			return
		}

		if cfp.currName != "" {
			for i := range cfp.Data.Contexts {
				if cfp.Data.Contexts[i].Name == cfp.currName {
					cfp.Data.Contexts[i].Name = name
					cfp.Data.Contexts[i].CtxData.URL = url
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
	cfp.form.GetFormItemByLabel("Description").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("URL").(*tview.InputField).SetText("nats://")
	cfp.form.GetFormItemByLabel("Token").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("User").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Password").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Creds").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Nkey").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Cert").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Key").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("CA").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("NSC").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Jetstream Domain").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Jetstream API Prefix").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Jetstream Event Prefix").(*tview.InputField).SetText("")
	cfp.form.GetFormItemByLabel("Inbox Prefix").(*tview.InputField).SetText("")

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

func createContextForm(ctx *ds.Context) *tview.Form {                                                                                     
	form := tview.NewForm()                                                                                                                      
	form.SetTitle("Context Form").SetBorder(true)                                                                                                
	form.AddInputField("Name", ctx.Name, 0, nil, nil)                                                                              
															
	ctxData := ctx.CtxData
	form.AddInputField("Description", ctxData.Description, 0, nil, nil)                                                                              
	form.AddInputField("URL", ctxData.URL, 0, nil, nil)                                                                                              
	form.AddInputField("Token", ctxData.Token, 0, nil, nil)                                                                                          
	form.AddInputField("User", ctxData.User, 0, nil, nil)                                                                                            
	form.AddInputField("Password", ctxData.Password, 0, nil, nil)                                                                                    
	form.AddInputField("Creds", ctxData.Creds, 0, nil, nil)                                                                                          
	form.AddInputField("Nkey", ctxData.Nkey, 0, nil, nil)                                                                                            
	form.AddInputField("Cert", ctxData.Cert, 0, nil, nil)                                                                                            
	form.AddInputField("Key", ctxData.Key, 0, nil, nil)                                                                                              
	form.AddInputField("CA", ctxData.CA, 0, nil, nil)                                                                                                
	form.AddInputField("NSC", ctxData.NSC, 0, nil, nil)                                                                                              
	form.AddInputField("Jetstream Domain", ctxData.JetstreamDomain, 0, nil, nil)                                                                     
	form.AddInputField("Jetstream API Prefix", ctxData.JetstreamAPIPrefix, 0, nil, nil)                                                              
	form.AddInputField("Jetstream Event Prefix", ctxData.JetstreamEventPrefix, 0, nil, nil)                                                          
	form.AddInputField("Inbox Prefix", ctxData.InboxPrefix, 0, nil, nil) 
	return form                                                                                                                                  
}  


func (cp *ContextFormPage) notify(message string, duration time.Duration, logLevel string) {
	cp.footerTxt.SetText(message)
	cp.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		cp.footerTxt.SetText("")
		cp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}
