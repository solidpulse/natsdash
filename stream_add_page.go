package main

import (
	"encoding/json"
	"time"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/nats-io/nats.go"
	"github.com/rivo/tview"
	"github.com/solidpulse/natsdash/ds"
	"github.com/solidpulse/natsdash/natsutil"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

type StreamAddPage struct {
	*tview.Flex
	app       *tview.Application
	Data      *ds.Data
	textArea  *tview.TextArea
	footerTxt *tview.TextView
	isEdit    bool
	streamName string
}

func NewStreamAddPage(app *tview.Application, data *ds.Data) *StreamAddPage {
	sap := &StreamAddPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
		Data: data,
	}

	sap.setupUI()
	sap.setupInputCapture()
	return sap
}

func (sap *StreamAddPage) setEditMode(name string) {
	sap.isEdit = true
	sap.streamName = name
}

func (sap *StreamAddPage) setupUI() {
	// Header
	headerRow := tview.NewFlex()
	headerRow.SetDirection(tview.FlexColumn)
	headerRow.SetBorderPadding(1, 0, 1, 1)

	headerRow.AddItem(createTextView("[ESC] Back", tcell.ColorWhite), 0, 1, false)
	headerRow.AddItem(createTextView("[Alt+Enter] Save", tcell.ColorWhite), 0, 1, false)
	headerRow.SetTitle("STREAM CONFIGURATION")
	sap.AddItem(headerRow, 3, 1, false)

	// Text area for YAML
	sap.textArea = tview.NewTextArea()
	sap.textArea.SetBorder(true)
	sap.textArea.SetTitle("Stream Configuration (YAML)")
	sap.AddItem(sap.textArea, 0, 1, true)

	// Footer
	footer := tview.NewFlex()
	footer.SetBorder(true)
	sap.footerTxt = createTextView("", tcell.ColorWhite)
	footer.AddItem(sap.footerTxt, 0, 1, false)
	sap.AddItem(footer, 3, 1, false)

	userFriendlyJSON5 := `{
    // Name of the stream (required)
    name: "my_stream",

    // Description of the stream (optional)
    description: "My Stream Description",

    // Subjects that messages can be published to (required)
    // Examples: ["orders.*", "shipping.>", "customer.orders.*"]
    subjects: [
        "my.subject.>"
    ],

    // Storage backend (required)
    // Possible values: "file", "memory"
    storage: "file",

    // Number of replicas for the stream
    // Range: 1-5
    num_replicas: 1,

    // Retention policy (required)
    // Possible values: "limits", "interest", "workqueue"
    retention: "limits",

    // Discard policy when limits are reached
    // Possible values: "old", "new"
    discard: "old",

    // Maximum number of messages in the stream
    // -1 for unlimited
    max_msgs: -1,

    // Maximum number of bytes in the stream
    // -1 for unlimited
    max_bytes: -1,

    // Maximum age of messages
    // Examples: "24h", "7d", "1y"
    max_age: "24h",

    // Maximum message size in bytes
    // -1 for unlimited
    max_msg_size: -1,

    // Maximum number of messages per subject
    // -1 for unlimited
    max_msgs_per_subject: -1,

    // Maximum number of consumers
    // -1 for unlimited
    max_consumers: -1
}`
	sap.textArea.SetText(userFriendlyJSON5, false)
}

func (sap *StreamAddPage) setupInputCapture() {
	sap.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			sap.notify("Loading......", 3*time.Second, "info")
			sap.goBack()
			return nil
		}
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			yamlText := sap.textArea.GetText()
			
			// Parse JSON5 into a map
			var jsonData map[string]interface{}
			err := json5.Unmarshal([]byte(yamlText), &jsonData)
			if err != nil {
				sap.notify("Invalid JSON5 configuration: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Convert to regular JSON
			jsonBytes, err := json.Marshal(jsonData)
			if err != nil {
				sap.notify("Error converting configuration: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// First unmarshal into an intermediate struct that keeps max_age as string
			var intermediateConfig struct {
				Name              string   `json:"name"`
				Description       string   `json:"description"`
				Subjects         []string `json:"subjects"`
				Retention        string   `json:"retention"`
				MaxConsumers     int      `json:"max_consumers"`
				MaxMsgs          int64    `json:"max_msgs"`
				MaxBytes         int64    `json:"max_bytes"`
				Discard          string   `json:"discard"`
				MaxAge           string   `json:"max_age"`
				MaxMsgsPerSubject int64    `json:"max_msgs_per_subject"`
				MaxMsgSize       int32    `json:"max_msg_size"`
				Storage          string   `json:"storage"`
				Replicas         int      `json:"num_replicas"`
			}
			err = json.Unmarshal(jsonBytes, &intermediateConfig)
			if err != nil {
				sap.notify("Invalid configuration: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Parse duration string
			maxAge, err := time.ParseDuration(intermediateConfig.MaxAge)
			if err != nil {
				sap.notify("Invalid max_age duration: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Convert string values to proper NATS types
			retention, err := parseRetentionPolicy(intermediateConfig.Retention)
			if err != nil {
				sap.notify("Invalid retention policy: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			storage, err := parseStorageType(intermediateConfig.Storage)
			if err != nil {
				sap.notify("Invalid storage type: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			discard, err := parseDiscardPolicy(intermediateConfig.Discard)
			if err != nil {
				sap.notify("Invalid discard policy: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Create final config
			config := nats.StreamConfig{
				Name:              intermediateConfig.Name,
				Description:       intermediateConfig.Description,
				Subjects:         intermediateConfig.Subjects,
				Retention:        retention,
				MaxConsumers:     intermediateConfig.MaxConsumers,
				MaxMsgs:          intermediateConfig.MaxMsgs,
				MaxBytes:         intermediateConfig.MaxBytes,
				Discard:          discard,
				MaxAge:           maxAge,
				MaxMsgsPerSubject: intermediateConfig.MaxMsgsPerSubject,
				MaxMsgSize:       intermediateConfig.MaxMsgSize,
				Storage:          storage,
				Replicas:         intermediateConfig.Replicas,
			}
			if err != nil {
				sap.notify("Invalid YAML configuration: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Connect to NATS
			conn, err := natsutil.Connect(&sap.Data.CurrCtx.CtxData)
			if err != nil {
				sap.notify("Failed to connect to NATS: "+err.Error(), 3*time.Second, "error")
				return nil
			}
			defer conn.Close()

			// Get JetStream context
			js, err := conn.JetStream()
			if err != nil {
				sap.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			// Convert to nats.StreamConfig
			streamConfig := nats.StreamConfig{
				Name:              config.Name,
				Description:       config.Description,
				Subjects:         config.Subjects,
				Retention:        config.Retention,
				MaxConsumers:     config.MaxConsumers,
				MaxMsgs:          config.MaxMsgs,
				MaxBytes:         config.MaxBytes,
				Discard:          config.Discard,
				MaxAge:           config.MaxAge,
				MaxMsgsPerSubject: config.MaxMsgsPerSubject,
				MaxMsgSize:       config.MaxMsgSize,
				Storage:          config.Storage,
				Replicas:         config.Replicas,
			}

			var err error
			if sap.isEdit {
				_, err = js.UpdateStream(&streamConfig)
			} else {
				_, err = js.AddStream(&streamConfig)
			}
			if err != nil {
				sap.notify("Failed to create stream: "+err.Error(), 3*time.Second, "error")
				return nil
			}

			sap.notify("Stream created successfully", 3*time.Second, "info")
			sap.goBack()
			return nil
		}
		return event
	})
}

func (sap *StreamAddPage) goBack() {
	pages.SwitchToPage("streamListPage")
	_, b := pages.GetFrontPage()
	b.(*StreamListPage).redraw(&sap.Data.CurrCtx)
	sap.app.SetFocus(b)
}

func (sap *StreamAddPage) redraw(ctx *ds.Context) {
	if !sap.isEdit {
		return
	}

	// Connect to NATS
	conn, err := natsutil.Connect(&ctx.CtxData)
	if err != nil {
		sap.notify("Failed to connect to NATS: "+err.Error(), 3*time.Second, "error")
		return
	}
	defer conn.Close()

	// Get JetStream context
	js, err := conn.JetStream()
	if err != nil {
		sap.notify("Failed to get JetStream context: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Get stream info
	stream, err := js.StreamInfo(sap.streamName)
	if err != nil {
		sap.notify("Failed to get stream info: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Convert the config to our intermediate format
	config := struct {
		Name              string   `json:"name"`
		Description       string   `json:"description"`
		Subjects         []string `json:"subjects"`
		Retention        string   `json:"retention"`
		MaxConsumers     int      `json:"max_consumers"`
		MaxMsgs          int64    `json:"max_msgs"`
		MaxBytes         int64    `json:"max_bytes"`
		Discard          string   `json:"discard"`
		MaxAge           string   `json:"max_age"`
		MaxMsgsPerSubject int64    `json:"max_msgs_per_subject"`
		MaxMsgSize       int32    `json:"max_msg_size"`
		Storage          string   `json:"storage"`
		Replicas         int      `json:"num_replicas"`
	}{
		Name:              stream.Config.Name,
		Description:       stream.Config.Description,
		Subjects:         stream.Config.Subjects,
		Retention:        retentionPolicyToString(stream.Config.Retention),
		MaxConsumers:     stream.Config.MaxConsumers,
		MaxMsgs:          stream.Config.MaxMsgs,
		MaxBytes:         stream.Config.MaxBytes,
		Discard:          discardPolicyToString(stream.Config.Discard),
		MaxAge:           stream.Config.MaxAge.String(),
		MaxMsgsPerSubject: stream.Config.MaxMsgsPerSubject,
		MaxMsgSize:       stream.Config.MaxMsgSize,
		Storage:          storageTypeToString(stream.Config.Storage),
		Replicas:         stream.Config.Replicas,
	}

	// Convert to JSON5 format with comments
	configJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		sap.notify("Failed to marshal config: "+err.Error(), 3*time.Second, "error")
		return
	}

	// Add comments to the JSON
	configWithComments := fmt.Sprintf(`{
    // Name of the stream (required)
    name: %q,

    // Description of the stream (optional)
    description: %q,

    // Subjects that messages can be published to (required)
    // Examples: ["orders.*", "shipping.>", "customer.orders.*"]
    subjects: %s,

    // Storage backend (required)
    // Possible values: "file", "memory"
    storage: %q,

    // Number of replicas for the stream
    // Range: 1-5
    num_replicas: %d,

    // Retention policy (required)
    // Possible values: "limits", "interest", "workqueue"
    retention: %q,

    // Discard policy when limits are reached
    // Possible values: "old", "new"
    discard: %q,

    // Maximum number of messages in the stream
    // -1 for unlimited
    max_msgs: %d,

    // Maximum number of bytes in the stream
    // -1 for unlimited
    max_bytes: %d,

    // Maximum age of messages
    // Examples: "24h", "7d", "1y"
    max_age: %q,

    // Maximum message size in bytes
    // -1 for unlimited
    max_msg_size: %d,

    // Maximum number of messages per subject
    // -1 for unlimited
    max_msgs_per_subject: %d,

    // Maximum number of consumers
    // -1 for unlimited
    max_consumers: %d
}`,
		config.Name,
		config.Description,
		prettyPrintSubjects(config.Subjects),
		config.Storage,
		config.Replicas,
		config.Retention,
		config.Discard,
		config.MaxMsgs,
		config.MaxBytes,
		config.MaxAge,
		config.MaxMsgSize,
		config.MaxMsgsPerSubject,
		config.MaxConsumers,
	)

	sap.textArea.SetText(configWithComments, true)

}

func (sap *StreamAddPage) notify(message string, duration time.Duration, logLevel string) {
	sap.footerTxt.SetText(message)
	sap.footerTxt.SetTextColor(getLogLevelColor(logLevel))

	go func() {
		time.Sleep(duration)
		sap.footerTxt.SetText("")
		sap.footerTxt.SetTextColor(tcell.ColorWhite)
	}()
}



func parseRetentionPolicy(s string) (nats.RetentionPolicy, error) {
	switch s {
	case "limits":
		return nats.LimitsPolicy, nil
	case "interest":
		return nats.InterestPolicy, nil
	case "workqueue":
		return nats.WorkQueuePolicy, nil
	default:
		return 0, fmt.Errorf("unknown retention policy: %s", s)
	}
}

func parseStorageType(s string) (nats.StorageType, error) {
	switch s {
	case "file":
		return nats.FileStorage, nil
	case "memory":
		return nats.MemoryStorage, nil
	default:
		return 0, fmt.Errorf("unknown storage type: %s", s)
	}
}

func parseDiscardPolicy(s string) (nats.DiscardPolicy, error) {
	switch s {
	case "old":
		return nats.DiscardOld, nil
	case "new":
		return nats.DiscardNew, nil
	default:
		return 0, fmt.Errorf("unknown discard policy: %s", s)
	}
}
func retentionPolicyToString(p nats.RetentionPolicy) string {
	switch p {
	case nats.LimitsPolicy:
		return "limits"
	case nats.InterestPolicy:
		return "interest"
	case nats.WorkQueuePolicy:
		return "workqueue"
	default:
		return "unknown"
	}
}

func storageTypeToString(s nats.StorageType) string {
	switch s {
	case nats.FileStorage:
		return "file"
	case nats.MemoryStorage:
		return "memory"
	default:
		return "unknown"
	}
}

func discardPolicyToString(d nats.DiscardPolicy) string {
	switch d {
	case nats.DiscardOld:
		return "old"
	case nats.DiscardNew:
		return "new"
	default:
		return "unknown"
	}
}

func prettyPrintSubjects(subjects []string) string {
	if len(subjects) == 0 {
		return "[]"
	}
	if len(subjects) == 1 {
		return fmt.Sprintf("[\n        %q\n    ]", subjects[0])
	}
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, subject := range subjects {
		sb.WriteString(fmt.Sprintf("        %q", subject))
		if i < len(subjects)-1 {
			sb.WriteString(",\n")
		}
	}
	sb.WriteString("\n    ]")
	return sb.String()
}
