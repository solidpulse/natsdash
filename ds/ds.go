package ds

type Data struct {
	//list of contexts
	Contexts []Context
	CurrCtx  Context
}

type Context struct {
	UUID string
	Name string
	URL  string
}

type ConfigData struct {
	//list of contexts
	Contexts []ConfigContext
}

type ConfigContext struct {
	UUID string
	Name string
	URL  string
}

// function to convert Data to ConfigData
func (data *Data) ToConfigData() *ConfigData {
	configData := &ConfigData{}
	for _, context := range data.Contexts {
		configData.Contexts = append(configData.Contexts, ConfigContext{UUID: context.UUID, Name: context.Name, URL: context.URL})
	}
	return configData
}

// function to convert ConfigData to Data, do a merge based on UUID
func (configData *ConfigData) refreshFromConfig(data *Data) {
	for _, context := range configData.Contexts {
		//check if context already exists
		for _, existingContext := range data.Contexts {
			if context.UUID == existingContext.UUID {
				existingContext.Name = context.Name
				existingContext.URL = context.URL
			}
		}
		//if not, add it
		data.Contexts = append(data.Contexts, Context{UUID: context.UUID, Name: context.Name, URL: context.URL})

	}
}
