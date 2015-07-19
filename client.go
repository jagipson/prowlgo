// Package prowlgo provides an interface to Prowl it allows to send push notifications to your
// iOS device. This package supports the complete prowl API including verify, retrieve/token and
// retrieve/apikey requests. For more info about the Prowl API see: http://www.prowlapp.com/api.php
//
// To use this package create a Client instance either by using NewClient or by creating the instance
// using the Builder.
package prowlgo

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	//PrioEmergency is the highest priority for emergenca messages
	PrioEmergency = 2
	//PrioHigh is a higher then normal priority
	PrioHigh = 1
	//PrioNormal is for regular messages
	PrioNormal = 0
	//PrioModerate is a lower then normal priority
	PrioModerate = -1
	//PrioVeryLow is the lowest priority for unimportant messages
	PrioVeryLow = -2
)

const (
	prowlBase         = "https://api.prowlapp.com/publicapi/"
	addURL            = prowlBase + "add"
	verifyURL         = prowlBase + "verify"
	retrieveTokenURL  = prowlBase + "retrieve/token"
	retrieveAPIKeyURL = prowlBase + "retrieve/apikey"

	defaultTimeout = 30 * time.Second
	waitSync       = -1 * time.Second

	defaultToProwlLabel = "(copied to prowl)"

	enter = true
	leave = false
)

// Client represents the prowl client which is used
// to send notifications to iOS devices running the prowl app via the prowl
// http api which is described here http://www.prowlapp.com/api.php
//
// With the client you can do the following
//  - Send messages to iOS devices (see Add)
//  - Retrieve a new api key for your application. The process here is
//    RetrieveToken, present approve URL to used and let them approve your request on the prowl website,
//    then RetrieveAPIKey and work with the new key.
//  - Write to the log and inparallel notify your iOS device (see Log and LogSync)
type Client struct {
	config       Config
	apiKeys      map[string]bool
	apiKeysDirty bool
	lock         chan bool
	unauthorized bool
	remaining    int
	reset        time.Time
}

// Config can be used to create a new Client. It might be handy if you need to
// persist your client for later use. See the Config() method for an example.
type Config struct {
	//APIKey is the api key that is to be used by this app. Needed for Add requests.
	//Not needed for RetrieveToken and RetrieveAPIKey calls.
	//The program is not working on this string array it is only there for simplified
	//configuration of the client. The client operates on a map which is copied back
	//when the Config() getter is called.
	APIKeys []string

	//ProviderKey is the provider key that is used in RetrieveToken and RetrieveAPIKey calls.
	//It must be defined for these calls. It is optional for calls to Add. Here it might
	//be usefull if prowl granted a higher api limit to the provider key.
	ProviderKey string

	//Token is received during RetrieveToken and must then be submitted during RetrieveAPIKey.
	//You will only need to fill in this field if you persisted the state of your
	//program and then shut it down in between those two calls. Then you need to let the
	//new instance know which token to use when continuing the process of retrieving an api key.
	Token string

	//Application is the application name which is displayed in the prowl app
	//on top of all messages sent by this client.
	Application string

	//The logger used by Log and LogSync. The standard logger will be used if nothing is defined here.
	Logger *log.Logger `json:"-"`

	//ToProwlLabel is the string label that is concatenated to the message that is written to
	//the log when the message is also sent to prowl. Client.Loc() and Client.LogSync() will append
	//this label to the message that goes to the log to indicate that the message was also sent
	//to the prowl app.
	ToProwlLabel *string
}

// Response represents the prowl server responses.
// The prowl server answers with a XML document which is parsed into this struct.
// Only parts of the struct will be filled with values depending on the the
// type of the response from the prowl server
//
// It's exported to make the xml unmarshalling work. It shouldn't be needed
// when you code against this API.
type Response struct {
	XMLName xml.Name `xml:"prowl"`
	Error   struct {
		XMLName xml.Name `xml:"error"`
		Code    int      `xml:"code,attr"`
		Message string   `xml:",chardata"`
	}
	Success struct {
		XMLName   xml.Name `xml:"success"`
		Code      int      `xml:"code,attr"`
		Remaining int      `xml:"remaining,attr"`
		Resetdate int64    `xml:"resetdate,attr"`
	}
	Retrieve struct {
		XMLName xml.Name `xml:"retrieve"`
		APIKey  string   `xml:"apikey,attr"`
		Token   string   `xml:"token,attr"`
		URL     string   `xml:"url,attr"`
	}
}

// NewClient creates a new Client from the provided config. The config can be partially empty.
// E.g. to send Add requests an api key and the application string
// will be enough. On the other hand the provider key will be sufficient to go through the process
// of retrieving a new api key. On the other hand ,
//
// Alternatively you can use the Builder to create a new Client instance.
func NewClient(config Config) (clt *Client, err error) {
	if config.APIKeys == nil {
		config.APIKeys = make([]string, 0)
	}
	if len(config.ProviderKey) != 40 && len(config.ProviderKey) != 0 {
		return nil, fmt.Errorf("provider key must either be 40 chars long or undefined")
	}
	if len(config.Token) != 40 && len(config.Token) != 0 {
		return nil, fmt.Errorf("token must either be 40 chars long or undefined")
	}
	if len(config.Application) > 256 {
		return nil, fmt.Errorf("application must not exceed 256 chars in length")
	}

	apiKeys := make(map[string]bool)
	for _, key := range config.APIKeys {
		if len(key) != 40 {
			return nil, fmt.Errorf("api key must either be 40 chars long or undefined")
		}
		if !apiKeys[key] {
			apiKeys[key] = true
		}
	}

	if config.Logger == nil {
		config.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	if config.ToProwlLabel == nil {
		cpy := defaultToProwlLabel
		config.ToProwlLabel = &cpy
	}

	return &Client{
		config:       config,
		apiKeys:      apiKeys,
		apiKeysDirty: len(apiKeys) != len(config.APIKeys),
		lock:         make(chan bool, 1),
		unauthorized: false,
		remaining:    1000,
		reset:        time.Now(),
	}, nil
}

// Add adds an event to the prowl queue which will be delivered to the client app
// as soon as possible. It takes a priority in the range of -2 (lowest) to 2 (highest)
// and an event string, which represents the title displayed at the client. description
// is the message body. Add will return an error if the message can't be delivered to
// the prowl server or in case of illegal arguments. If the request was successful
// the number of remaining prowl api requests (messages that can be sent) will be returned.
func (clt *Client) Add(priority int, event string, description string) (remaining int, err error) {
	return clt.AddWithURL(priority, event, description, "")
}

// AddWithURL is the same as Add() with an additional URL argument. This URL will be presented to the
// user of the prowl app. It can be tapped to open the respective location.
// The URL will be appended to the description and also be sent as an argument to the prowl app.
// As of the writing of this code the prowl app displays a little (i) next to the message that
// the user can tap to open the URL.
func (clt *Client) AddWithURL(priority int, event string, description string, withURL string) (remaining int, err error) {
	if clt.unauthorized {
		return clt.remaining, fmt.Errorf("the api key is know to be invalid")
	}
	if len(clt.config.APIKeys) == 0 {
		return clt.remaining, fmt.Errorf("a valid api key is required for add operation")
	}
	if priority < -2 || priority > 2 {
		return clt.remaining, fmt.Errorf("priority argument must be in the range -2..2")
	}
	if len(event) > 1024 {
		return clt.remaining, fmt.Errorf("event argument must not exceed 1024 chars")
	}
	if len(description) > 10000 {
		return clt.remaining, fmt.Errorf("description argument must not exceed 10000 chars")
	}
	if len(withURL) > 256 {
		return clt.remaining, fmt.Errorf("withURL argument must not exceed 512 chars")
	}
	if clt.remaining <= 0 && clt.reset.After(time.Now()) {
		return clt.remaining, fmt.Errorf("api requests spent; come back after %s", clt.reset)
	}

	event = strings.TrimSpace(event)
	description = strings.TrimSpace(description)
	withURL = strings.TrimSpace(withURL)

	if len(withURL) > 0 {
		if len(description)+len(withURL)+4 > 10000 {
			description = strings.TrimSpace(description[0:10000-len(withURL)-4]) + "..."
		}
		description = description + " " + withURL
	}

	resp, err := http.PostForm(addURL, url.Values{
		"apikey":      {clt.makeAPIKeyRequestArgument()},
		"providerkey": {clt.config.ProviderKey},
		"priority":    {fmt.Sprintf("%d", priority)},
		"application": {clt.config.Application},
		"event":       {event},
		"description": {description},
		"url":         {withURL},
	})
	response, err := clt.handleResponse(resp, err)

	if err != nil {
		return clt.remaining, fmt.Errorf("add request to prowl server failed: %s", err)
	}

	if response.Success.XMLName.Local != "" {
		clt.reset = time.Unix(response.Success.Resetdate, 0)
		clt.remaining = response.Success.Remaining
	}

	return clt.remaining, nil
}

// Verify verifys the validity of the provided api key.
// Invoking this method will cost you an prowl api call.
// If you have a provider key which is whitlisted for a higher call limit make
// sure you use a client which is configured with that provider key.
// If the key is not valid it will return an error. If the key is OK
// the number of remaining api calls is returned.
func (clt *Client) Verify(apiKey string) (remaining int, err error) {
	remaining = clt.remaining
	if len(apiKey) != 40 {
		err = fmt.Errorf("apiKey argument must be exactly 40 chars long")
		return
	}

	u, err := url.Parse(verifyURL)
	if err != nil {
		panic("verify url can not be parsed")
	}

	q := u.Query()
	q.Set("apikey", apiKey)
	if len(clt.config.ProviderKey) == 40 {
		q.Set("providerkey", clt.config.ProviderKey)
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	_, err = clt.handleResponse(resp, err)
	if err != nil {
		return clt.remaining, fmt.Errorf("verify request to prowl server failed: %s", err)
	}

	return clt.remaining, nil
}

// RetrieveToken retrieves a token from the prowl server. This token has to
// be approved by the used using the returned approveURL. Once this has happend
// a call to RetrieveAPIKey can be made to retrieve a new api key from the server.
//
// This call requires that a provider key was configured when thie client was
// created. Other config is not necessary to retrieve a token.
//
// If this client does not remains running while the user approves the request
// make sure to persist the config of this client and create the new client with
// the persisted config later on. The config also contains the token (together with
// the provider key that you also will need during RetrieveAPIKey)
func (clt *Client) RetrieveToken() (approveURL string, err error) {
	if len(clt.config.ProviderKey) != 40 {
		err = fmt.Errorf("provider key is required for retrieve token operation")
		return
	}

	u, err := url.Parse(retrieveTokenURL)
	if err != nil {
		panic("retrieve token url can not be parsed")
	}

	q := u.Query()
	q.Set("providerkey", clt.config.ProviderKey)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	response, err := clt.handleResponse(resp, err)
	if err != nil {
		err = fmt.Errorf("retrieve token request to prowl server failed: %s", err)
		return
	}

	approveURL = response.Retrieve.URL
	clt.config.Token = response.Retrieve.Token

	return
}

// RetrieveAPIKey retrieves a new api key from the prowl server. This call requires
// that this client is configured with a valied provider key and a vaild token.
//
//For an Example see Client.RetrieveToken
func (clt *Client) RetrieveAPIKey() (apiKey string, err error) {
	if len(clt.config.Token) != 40 {
		err = fmt.Errorf("token is required for retrieve token operation")
		return
	}
	if len(clt.config.ProviderKey) != 40 {
		err = fmt.Errorf("provider key is required for retrieve token operation")
		return
	}

	u, err := url.Parse(retrieveAPIKeyURL)
	if err != nil {
		panic("retrieve api key url can not be parsed")
	}

	q := u.Query()
	q.Set("providerkey", clt.config.ProviderKey)
	q.Set("token", clt.config.Token)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	response, err := clt.handleResponse(resp, err)
	if err != nil {
		err = fmt.Errorf("retrieve api key request to prowl server failed: %s", err)
		return
	}

	clt.apiKeys[apiKey] = true

	return response.Retrieve.APIKey, nil
}

func (clt *Client) makeAPIKeyRequestArgument() (req string) {
	i := 0
	for key := range clt.apiKeys {
		req += key
		if i < len(clt.config.APIKeys)-1 {
			req += ","
		}
		i++
	}
	return
}

func (clt *Client) handleResponse(resp *http.Response, inerr error) (response Response, err error) {
	if inerr != nil {
		err = fmt.Errorf("HTTP request to prowl server failed: %s", inerr)
		return
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			cerr = fmt.Errorf("error closing HTTP response body: %s", cerr)
			fmt.Println(cerr)
			if err == nil {
				err = cerr
			} else {
				err = fmt.Errorf("%s. Followed by: %s", err, cerr)
			}
		}
	}()

	buf := make([]byte, 1024)
	if _, err = resp.Body.Read(buf); err != nil {
		err = fmt.Errorf("can't read HTTP response body: %s", err)
		return
	}

	err = xml.Unmarshal(buf, &response)
	if err != nil {
		err = fmt.Errorf("can't unmarshal xml response from prowl server: %s", err)
		return
	}

	if len(response.Error.XMLName.Local) != 0 {
		//BUG: unauthorized logic must only apply to Add calls!
		if response.Error.Code == 401 {
			clt.unauthorized = true
		}
		err = fmt.Errorf("prowl returned error code %d: %s", response.Error.Code, response.Error.Message)
		return
	}

	if len(response.Success.XMLName.Local) != 0 {
		clt.reset = time.Unix(response.Success.Resetdate, 0)
		clt.remaining = response.Success.Remaining
	}

	return
}

// Config returns the config of this client. This can be handy if you need to
// persist the provider key and the token recieved from a call to RetrieveToken.
// It might take some time until the user approves the request and another client instance
// might be processing the following step (RetrieveAPIKey). This new instance can
// easily be configured with the Config returned by this call.
//
//You might consider persisting it in a JSON serialization. This will omit the Config.Logger field
//automatically.
func (clt *Client) Config() Config {
	clt.mutex(enter)
	defer clt.mutex(leave)

	if clt.apiKeysDirty {
		clt.config.APIKeys = make([]string, len(clt.apiKeys))
		i := 0
		for key := range clt.apiKeys {
			clt.config.APIKeys[i] = key
			i++
		}
	}
	clt.apiKeysDirty = false

	return clt.config
}

func (clt *Client) logWait(prio int, event string, message string, wait time.Duration) {
	clt.config.Logger.Println(event + ": " + message + " " + *clt.config.ToProwlLabel)
	sent := make(chan bool, 1)
	prowlf := func() {
		if _, err := clt.Add(prio, event, message); err != nil {
			if len(message) > 20 {
				message = strings.TrimSpace(message[0:17]) + "..."
			}
			if len(event) > 10 {
				event = strings.TrimSpace(event[0:7]) + "..."
			}
			clt.config.Logger.Printf("can't send prowl message (\"%s: %s\") %s", event, message, err)
		}
		sent <- true
	}

	if wait == waitSync {
		prowlf()
	} else {
		go prowlf()
		select {
		case <-time.After(wait):
			clt.config.Logger.Println("Timeout while sending prowl message")
		case <-sent:
		}
	}
}

// Log is shorthand for writing the event and description to the configured logger and
// concurrently sending the message to the prowl server. The call will report an error
// in the logs if sending to the server fails or times out.
func (clt *Client) Log(prio int, event string, message string) {
	clt.logWait(prio, event, message, defaultTimeout)
}

// LogSync performs the same actions as Log but will block until the request to the prowl
// server returns.
func (clt *Client) LogSync(prio int, event string, message string) {
	clt.logWait(prio, event, message, waitSync)

}

func (clt *Client) String() string {
	return fmt.Sprintf("prowl client for application %s, %d api requests left, reset at %s", clt.config.Application, clt.remaining, clt.reset)
}

func (clt *Client) mutex(enter bool) {
	if enter {
		clt.lock <- true
	} else {
		<-clt.lock
	}
}
