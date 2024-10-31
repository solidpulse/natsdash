package ds

import (
	"encoding/json"
	"os"

	"path/filepath"
	"strings"

	"github.com/evnix/natsdash/logger"
	"github.com/nats-io/nats.go"
)

type Data struct {
	//list of contexts
	Contexts []Context
	CurrCtx  Context `json:"-"`
}

type NatsCliContext struct {
	Description string `json:"description"`
	URL         string `json:"url"`
	Token       string `json:"token"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Creds       string `json:"creds"`
	Nkey        string `json:"nkey"`
	Cert        string `json:"cert"`
	Key         string `json:"key"`
	CA          string `json:"ca"`
	NSC         string `json:"nsc"`
	JetstreamDomain       string `json:"jetstream_domain"`
	JetstreamAPIPrefix    string `json:"jetstream_api_prefix"`
	JetstreamEventPrefix  string `json:"jetstream_event_prefix"`
	InboxPrefix           string `json:"inbox_prefix"`
	UserJWT               string `json:"user_jwt"`
}
type Context struct {
	Name        string 
	CtxData     NatsCliContext 
	LogFilePath string      `json:"-"`
	LogFile     *os.File    `json:"-"`
	Conn        *nats.Conn `json:"-"`
	CoreNatsSub *nats.Subscription `json:"-"`
}


func GetConfigDir() (string, error) {
	// Get the user's home directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "",err
	}
	//create config directory if it doesn't exist
	configDir := userConfigDir + "/nats"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.Mkdir(configDir, 0755)
		if err != nil {
			return "",err
		}
	}
	return configDir, nil
}

// function to save ConfigData to multiple files in the users config directory
func (configData *Data) SaveToFile() error {

	//create config directory if it doesn't exist
	configDir, err := GetConfigDir()
	if err != nil {
		logger.Error("Failed to get config directory: %v", err) // Add this line
		return err
	}
	
	// Save each context to a separate file
	for _, context := range configData.Contexts {
		filePath := filepath.Join(configDir, context.Name+".json")
		file, err := os.Create(filePath)
		if err != nil {
			logger.Error("Failed to create file %s: %v", filePath, err) // Add this line
			return err
		}
		defer file.Close()

		err = json.NewEncoder(file).Encode(context.CtxData)
		if err != nil {
			logger.Error("Failed to encode context data to file %s: %v", filePath, err) // Add this line
			return err
		}
		logger.Info("Successfully saved context data to file %s", filePath) // Add this line
	}

	return nil
}



// function to load ConfigData from directory in users config directory
func (data *Data) LoadFromDir(dirPath string) error {
	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Read all files in the directory
	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	// Iterate over each file
	for _, file := range files {
		// Check if the file is a JSON file
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			// Open the file
			filePath := filepath.Join(dirPath, file.Name())
			logger.Info("Loading context data from file %s", filePath) // Add this line
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			// Unmarshal the file contents into a NatsCliContext
			var ctx NatsCliContext
			err = json.NewDecoder(file).Decode(&ctx)
			if err != nil {
				return err
			}

			// Create a new Context with the filename as the Name and the unmarshaled data
			ctxName := strings.Split(strings.TrimSuffix(file.Name(), ".json"), string(filepath.Separator))
			context := Context{
				Name:    ctxName[len(ctxName)-1],
				CtxData: ctx,
			}

			// Append the new Context to the Data's Contexts slice
			data.Contexts = append(data.Contexts, context)
		}
	}
	logger.Debug("Loaded %d contexts %s", len(data.Contexts), data.Contexts)

	return nil
}
