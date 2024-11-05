package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/logger"
	"github.com/solidpulse/natsdash/natsutil"
)

type ContextPage struct {
	*tview.Flex
	Data        *ds.Data
	ctxListView *tview.List
	app         *tview.Application // Add this line
	footerTxt   *tview.TextView
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
	cp.AddItem(headerRow, 4, 4, false)

	// Context list setup
	ctxListBox := tview.NewFlex()
	ctxListBox.SetTitle("Contexts").SetBorder(true)
	cp.ctxListView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	ctxListBox.AddItem(cp.ctxListView, 0, 20, false)
	ctxListBox.SetBorderPadding(0, 0, 1, 1)
	cp.AddItem(ctxListBox, 0, 18, false)

	// Footer setup
	footer := tview.NewFlex()
	footer.SetBorder(true)
	footer.SetDirection(tview.FlexRow)
	footer.SetBorderPadding(0, 0, 1, 1)

	cp.footerTxt = createTextView(" -- ", tcell.ColorWhite)
	go cp.displayLicenseCopyrightInfo()
	footer.AddItem(cp.footerTxt, 0, 1, false)
	cp.AddItem(footer, 3, 2, false)
	cp.SetBorderPadding(1, 0, 1, 1)
	// Read NATS CLI contexts
	cp.reloadNatsCliContexts()
	contexts := cp.Data.Contexts
	cp.Data.Contexts = contexts
	logger.Info("Contexts to be added: %s", (contexts))
	for _, ctx := range contexts {
		logger.Info("Adding context in list %s", ctx.Name)
		cp.ctxListView.AddItem(ctx.Name, "", 0, nil)
	}

}

func (cp *ContextPage) displayLicenseCopyrightInfo() {
	// Fetch the info.json content from the URL
	resp, err := http.Get("https://raw.githubusercontent.com/solidpulse/natsdash/refs/heads/master/info.env")
	if err != nil {
		cp.footerTxt.SetText("Error fetching info")
		return
	}
	defer resp.Body.Close()

	// Parse the JSON response
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cp.footerTxt.SetText("Error reading info")
		return
	}

	// Parse the .env content
	envContent := string(body)
	envMap := make(map[string]string)
	for _, line := range strings.Split(envContent, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, `"`) // Remove surrounding quotes
			envMap[key] = value
		}
	}

	// Extract the required fields
	message := envMap["message"]
	showVersion := envMap["show_version"] == "true"
	currentVersion := envMap["current_version"]
	isNotice := envMap["is_notice"] == "true"

	// Update the footer text with the fetched information
	currVersion := ds.Version
	footerText := message
	if showVersion {
		footerText = fmt.Sprintf("%s | Current: %s | Latest: %s", message, currVersion, currentVersion)
	}
	if isNotice {
		cp.footerTxt.SetTextColor(getLogLevelColor("warn"))
	}
	cp.footerTxt.SetText(footerText)
	go cp.app.Draw()
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

			//delete the context file
			err := cp.Data.RemoveContextFileByName(cp.Data.Contexts[idx].Name)
			if err != nil {
				cp.notify(fmt.Sprintf("Error deleting context file: %s", err.Error()), 5*time.Second, "error")
				return event
			}
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
		} else if event.Rune() == 'j' || event.Rune() == 'J' {
			idx := cp.ctxListView.GetCurrentItem()
			data.CurrCtx = cp.Data.Contexts[idx]
			pages.SwitchToPage("streamListPage")
			_, b := pages.GetFrontPage()
			b.(*StreamListPage).redraw(&data.CurrCtx)
		} else if event.Rune() == 'n' || event.Rune() == 'N' {
			idx := cp.ctxListView.GetCurrentItem()
			if len(cp.Data.Contexts) == 0 {
				cp.notify("No contexts available", 5*time.Second, "error")
				return event
			}
			data.CurrCtx = cp.Data.Contexts[idx]

			// Connect to NATS
			go func() {
				cp.notify("Connecting to NATS...", 5*time.Second, "info")
				conn, err := natsutil.Connect(&data.CurrCtx.CtxData)
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

func (cp *ContextPage) reloadNatsCliContexts() {
	configDir, _ := ds.GetConfigDir()
	data.LoadFromDir(configDir)
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
