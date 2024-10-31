package main

import (
	"fmt"
	"os"
	"path"
	"time"
	"encoding/json"

	"github.com/evnix/natsdash/ds"
	"github.com/evnix/natsdash/natsutil"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContextPage struct {
	*tview.Flex
	Data        *ds.Data
	ctxListView *tview.List
	app         *tview.Application // Add this line
	footerTxt      *tview.TextView
}

func NewContextPage(app *tview.Application, data *ds.Data) *ContextPage {
	cp := &ContextPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app, // Add this line
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
	footer := tview.NewFlex()
	footer.SetBorder(true)
	footer.SetBorderPadding(0, 0, 1, 1)
	cp.footerTxt = createTextView("Primary Author: Avinash D'Silva | contact: dsilva.avinash@outlook.com", tcell.ColorWhite)
	footer.AddItem(cp.footerTxt, 0, 1, false)
	cp.AddItem(footer, 0, 2, false)
	cp.SetBorderPadding(1, 0, 1, 1)
	// Read NATS CLI contexts
	contexts, err := cp.readNatsCliContexts()
	if err != nil {
		cp.notify(fmt.Sprintf("Error reading NATS CLI contexts: %s", err.Error()), 5*time.Second, "error")
	} else {
		for _, ctx := range contexts {
			cp.ctxListView.AddItem(ctx.Name, "", 0, nil)
		}
	}



}

func (cp *ContextPage) setupInputCapture() {
	cp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			cp.app.Stop() // Add this line
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			cp.app.SetFocus(cp.ctxListView) // Add this line
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
			if len(cp.Data.Contexts) == 0 {
				return event
			}
			data.CurrCtx = cp.Data.Contexts[idx]


			// Connect to NATS
			go func() {
				cp.notify("Connecting to NATS...", 5*time.Second, "info")
				conn, err := natsutil.Connect(data.CurrCtx.CtxData.URL)
				if err != nil {
					cp.notify(fmt.Sprintf("Error connecting to NATS: %s", err.Error()), 5*time.Second, "error")
					return
				}
				data.CurrCtx.Conn = conn

				// Open log file
				currentTime := time.Now().Format("2006-01-02")
				logFilePath := path.Join(os.TempDir(), "natsdash", fmt.Sprintf("%s_%s.log", currentTime, data.CurrCtx.Name[:4]))
				logDir := path.Dir(logFilePath)
				if err := os.MkdirAll(logDir, 0755); err != nil {
					cp.notify(fmt.Sprintf("Error creating log directory: %s", err.Error()), 5*time.Second, "error")
					return
				}
				logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
				if err != nil {
					cp.notify(fmt.Sprintf("Error opening log file: %s", err.Error()), 5*time.Second, "error")
					return
				}
				data.CurrCtx.LogFilePath = logFilePath
				data.CurrCtx.LogFile = logFile
				logFile.WriteString("Connected to NATS. ClusterName: " + conn.ConnectedClusterName() +
				" ServerID: " + conn.ConnectedServerId() + "\n")
				pages.SwitchToPage("natsPage")
				_, b := pages.GetFrontPage()
				b.(*NatsPage).redraw(&data.CurrCtx)

			}()
		}
		return event
	})
}

func (cp *ContextPage) notify(message string, duration time.Duration, logLevel string) {
	cp.footerTxt.SetText(message)
	cp.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		cp.footerTxt.SetText("")
		cp.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}



func (cp *ContextPage) readNatsCliContexts() ([]ds.Context, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	contextDir := path.Join(configDir, "nats", "context")
	files, err := os.ReadDir(contextDir)
	if err != nil {
		return nil, err
	}

	var contexts []ds.Context
	for _, file := range files {
		if file.IsDir() || path.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := path.Join(contextDir, file.Name())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var context ds.Context
		err = json.Unmarshal(fileContent, &context)
		if err != nil {
			return nil, err
		}

		contexts = append(contexts, context)
	}

	return contexts, nil
}

func (cp *ContextPage) Redraw() {
	cp.ctxListView.Clear()
	cp.footerTxt.SetText("")
	for _, ctx := range cp.Data.Contexts {
		cp.ctxListView.AddItem(ctx.Name, "", 0, nil)
	}
	go cp.app.Draw()
}

func createContexListHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(0, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createColoredTextView("[white:green] NATS [yellow:white] DASH ", tcell.ColorWhite), 0, 1, false)
	headerRow1.AddItem(createTextView("", tcell.ColorWhite), 0, 1, false)
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
