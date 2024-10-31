package ds

import (
	"encoding/json"
	"os"

	"github.com/nats-io/nats.go"
)

type Data struct {
	//list of contexts
	Contexts []Context
	CurrCtx  Context `json:"-"`
}

type Context struct {
	UUID string
	Name string
	URL  string
	Conn *nats.Conn `json:"-"`
}

// function to save ConfigData to file in users config directory
func (configData *Data) SaveToFile() error {
	//get user config directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	//create config directory if it doesn't exist
	configDir := userConfigDir + "/natsdash"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.Mkdir(configDir, 0755)
		if err != nil {
			return err
		}
	}
	//create config file
	configFile, err := os.Create(configDir + "/config.json")
	if err != nil {
		return err
	}
	defer configFile.Close()
	//write config data to file

	//write config data to file
	err = json.NewEncoder(configFile).Encode(configData)
	if err != nil {
		return err
	}

	return nil
}

// function to load ConfigData from file in users config directory
func (data *Data) LoadFromFile() error {
	//get user config directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	//create config directory if it doesn't exist
	configDir := userConfigDir + "/natsdash"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.Mkdir(configDir, 0755)
		if err != nil {
			return err
		}
	}
	//create config file
	configFile, err := os.Open(configDir + "/config.json")
	if err != nil {
		return err
	}
	defer configFile.Close()
	configData := &Data{}
	//read config data from file
	err = json.NewDecoder(configFile).Decode(configData)
	if err != nil {
		return err
	}

	//merge contexts if they exist based on UUID else append
	for _, ctx := range configData.Contexts {
		found := false
		for i, c := range data.Contexts {
			if c.UUID == ctx.UUID {
				data.Contexts[i] = ctx
				found = true
			}
		}
		if !found {
			data.Contexts = append(data.Contexts, ctx)
		}
	}

	return nil
}
