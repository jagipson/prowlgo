package prowlgo

import "log"

// Builder is used to create a new prowl client -- if you like the
// builder pattern better than providing a config struct to the NewClient function.
type Builder struct {
	config Config
}

// NewBuilder creates a new prowl client builder which can be used
// to create and cofigure a new prowl clients with a number if chained commands.
func NewBuilder() *Builder {
	return &Builder{}
}

// SetAPIKey defines the api key for the client that will be built using this builder. Needs to be
//called when your client needs to make Add requests.
func (bld *Builder) SetAPIKey(apiKey string) *Builder {
	bld.config.APIKey = apiKey
	return bld
}

// SetToken defines the token for the client that will be built using this builder.
// Needs to be set if this client should continue the process of retrieving a new api key
// after the user approved the request.
func (bld *Builder) SetToken(token string) *Builder {
	bld.config.Token = token
	return bld
}

// SetProviderKey defines the provider key for the client that will be built using this builder.
// A provider key is required when you plan to retrieve a new api key or when
// you want to profit from a higher api call limit set for the provider.
func (bld *Builder) SetProviderKey(providerKey string) *Builder {
	bld.config.ProviderKey = providerKey
	return bld
}

// SetApplication defines the application string for the client that will be built using this builder.
// Should be called when your client will make Add requests.
func (bld *Builder) SetApplication(application string) *Builder {
	bld.config.Application = application
	return bld
}

// SetLogger defines the logger that is to be used by Client.Log() and Client.LogSync()
func (bld *Builder) SetLogger(logger *log.Logger) *Builder {
	bld.config.Logger = logger
	return bld
}

// SetToProwlLabel defines the label that is appended to log messages which are also
// added to prowl. I.e. Client.Log() or Client.LogSync() will append this label to
// the log message when they write the message to the log
// that is also sent to prowl. Someone inspecting
// the logs will know that the message is also know the user of the prowl app.
func (bld *Builder) SetToProwlLabel(label string) *Builder {
	bld.config.ToProwlLabel = &label
	return bld
}

// Build creates and returns the new prowl client. If any of the previous calls provided
// illegal client configuration this call will raise the respective error.
func (bld *Builder) Build() (client *Client, err error) {
	return NewClient(bld.config)
}
