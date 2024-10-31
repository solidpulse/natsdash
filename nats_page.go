package main

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/evnix/natsdash/ds"
	"github.com/gdamore/tcell/v2"
	"github.com/nats-io/nats.go"
	"github.com/rivo/tview"
)

type NatsPage struct {
	*tview.Flex
	Data         *ds.Data
	app          *tview.Application // Add this line
	subjectFilter *tview.InputField
	logView       *tview.TextView
	subjectName   *tview.InputField
	txtArea       *tview.TextArea
	tailingDone  chan struct{} // Add this line
	tailingMutex sync.Mutex    // Add this line
}

func NewNatsPage(app *tview.Application, data *ds.Data) *NatsPage {
	cfp := &NatsPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app, // Add this line
		tailingDone: make(chan struct{}), // Add this line
	}
	cfp.Data = data
	cfp.setupUI()
	cfp.setupInputCapture()

	return cfp
}

func (cfp *NatsPage) setupUI() {
	// Header setup
	headerRow := createNatsPageHeaderRow()
	cfp.AddItem(headerRow, 0, 6, false)

	// Initialize fields
	cfp.subjectFilter = tview.NewInputField()
	cfp.subjectFilter.SetLabel("Filter Subjects: ")
	cfp.subjectFilter.SetBorder(true)
	cfp.subjectFilter.SetBorderPadding(0, 0, 1, 1)
	// cfp.subjectFilter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	if event.Key() == tcell.KeyTab {
	// 		cfp.app.SetFocus(cfp.subjectName)
	// 		return nil
	// 	}
	// 	return event
	// })
	cfp.subjectFilter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			cfp.app.SetFocus(cfp.subjectName)
			return nil
		}
		return event
	})
	cfp.subjectFilter.SetDoneFunc(func(key tcell.Key) {
		cfp.subscribeToSubject(cfp.subjectFilter.GetText())
		cfp.app.SetFocus(cfp.logView)
	})
	cfp.AddItem(cfp.subjectFilter, 0, 6, false)

	cfp.logView = tview.NewTextView()
	cfp.logView.SetTitle(cfp.Data.CurrCtx.LogFilePath)
	cfp.logView.SetBorder(true)
	cfp.logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			cfp.app.SetFocus(cfp.subjectName)
			return nil
		}
		return event
	})
	cfp.AddItem(cfp.logView, 0, 50, false)

	cfp.subjectName = tview.NewInputField()
	cfp.subjectName.SetLabel("Target Subject: ")
	cfp.subjectName.SetBorder(true)
	cfp.subjectName.SetDoneFunc(func(key tcell.Key) {
		cfp.app.SetFocus(cfp.txtArea)
	})
	cfp.subjectName.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			cfp.app.SetFocus(cfp.txtArea)
			return nil
		}
		return event
	})
	cfp.AddItem(cfp.subjectName, 0, 6, false)

	cfp.txtArea = tview.NewTextArea()
	cfp.txtArea.SetPlaceholder("Message...")
	cfp.txtArea.SetBorder(true)
	cfp.txtArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			cfp.app.SetFocus(cfp.subjectFilter)
			return nil
		} else if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			cfp.sendMessage()
			return nil
		}
		return event
	})
	cfp.AddItem(cfp.txtArea, 0, 8, false)
	cfp.SetBorderPadding(0, 0, 1, 1)
}

func (cfp *NatsPage) redraw(ctx *ds.Context) {
	// Update log view title with the current context's log file path
	cfp.logView.SetTitle(ctx.LogFilePath)
	cfp.resetTailFile(ctx.LogFilePath)
	cfp.app.SetFocus(cfp.subjectFilter)
	go cfp.app.Draw()
}

func (cfp *NatsPage) setupInputCapture() {
	cfp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc:
			cfp.goBackToContextPage()
			return nil
		}
		return event
	})
}

func (cfp *NatsPage) goBackToContextPage() {
	// Stop the tailing goroutine
	cfp.tailingMutex.Lock()
	if cfp.tailingDone != nil {
		close(cfp.tailingDone)
		cfp.tailingDone = nil
	}
	cfp.tailingMutex.Unlock()

	pages.SwitchToPage("contexts")
	_, b := pages.GetFrontPage()
	b.(*ContextPage).Redraw()
	cfp.app.SetFocus(b) // Add this line
}

func createNatsPageHeaderRow() *tview.Flex {
	headerRow := tview.NewFlex()
	headerRow.SetBorder(false)
	headerRow.
		SetDirection(tview.FlexColumn).
		SetBorderPadding(1, 0, 1, 1)

	headerRow1 := tview.NewFlex()
	headerRow1.SetDirection(tview.FlexRow)
	headerRow1.SetBorder(false)

	headerRow1.AddItem(createTextView("[Esc] Back  | [Alt+Enter] Send | [F2] Filter | [F3] Subject | [F4] Body", tcell.ColorWhite), 0, 1, false)

	headerRow.AddItem(headerRow1, 0, 1, false)
	headerRow.SetTitle("NATS-DASH")

	return headerRow
}
func (cfp *NatsPage) resetTailFile(logFilePath string) {
	// Stop the previous tailing goroutine
	cfp.tailingMutex.Lock()
	if cfp.tailingDone != nil {
		close(cfp.tailingDone)
	}
	cfp.tailingDone = make(chan struct{})
	cfp.tailingMutex.Unlock()

	// Clear the log view
	cfp.logView.Clear()

	// Open the log file


	// Tail the log file and update the log view
	go func() {
		logFile, err := os.Open(logFilePath)
		if err != nil {
			cfp.app.QueueUpdateDraw(func() {
				cfp.logView.Write([]byte("Error opening log file: " + err.Error() + "\n"))
			})
			return
		}
		defer logFile.Close()
		buf := make([]byte, 1024)
		offset, _ := logFile.Seek(0, io.SeekEnd)
		for {
			select {
			case <-cfp.tailingDone:
				return
			default:
				if cfp.tailingDone == nil {
					return
				}
				n, err := logFile.ReadAt(buf, offset)
				if err != nil && err != io.EOF {
					cfp.app.QueueUpdateDraw(func() {
						cfp.logView.Write([]byte("Error reading log file: " + err.Error() + "\n"))
					})
					return
				}
				if n > 0 {
					cfp.app.QueueUpdateDraw(func() {
						cfp.logView.Write(buf[:n])
					})
					offset += int64(n)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (cfp *NatsPage) subscribeToSubject(subject string) {
	hourMinSec := time.Now().Format("15:04:05")
    // Unsubscribe from the previous subject if any
    if cfp.Data.CurrCtx.CoreNatsSub != nil {
        cfp.Data.CurrCtx.CoreNatsSub.Unsubscribe()
    }

    // Subscribe to the new subject
    sub, err := cfp.Data.CurrCtx.Conn.Subscribe(subject, func(msg *nats.Msg) {
        // Log the incoming message to the log file
        
        cfp.Data.CurrCtx.LogFile.WriteString(hourMinSec + " SUB[" + msg.Subject + "] " + string(msg.Data) + "\n")
		cfp.logView.ScrollToEnd()
    })
	if err != nil {
		cfp.Data.CurrCtx.LogFile.WriteString(err.Error()+ "\n")
	}else{
		cfp.Data.CurrCtx.LogFile.WriteString(hourMinSec + " Subscribed to " + subject + "\n")
	}
	cfp.Data.CurrCtx.CoreNatsSub = sub
	
}

func (cfp *NatsPage) sendMessage() {
	// Implement the logic to send the message here
	// For example, you can get the message from cfp.txtArea and send it via NATS
	message := cfp.txtArea.GetText()
	subject := cfp.subjectName.GetText()
	cfp.Data.CurrCtx.Conn.Publish(subject, []byte(message))
	hourMinSec := time.Now().Format("15:04:05")
	cfp.Data.CurrCtx.LogFile.WriteString(hourMinSec+" PUB["+ subject + "] " + message + "\n")
	cfp.logView.ScrollToEnd()
}
