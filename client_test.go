package prowlgo_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	prowl "github.com/tweithoener/prowlgo"
)

func TestNewClient(t *testing.T) {
	_, err := prowl.NewClient(prowl.Config{})
	if err != nil {
		t.Error("error creating empty client")
	}

	_, err = prowl.NewClient(prowl.Config{
		Token: "12345",
	})
	if err == nil {
		t.Error("invalid token should produce an error")
	}

	_, err = prowl.NewClient(prowl.Config{
		Token:       "1234512345123451234512345123451234512345",
		ProviderKey: "12345",
	})
	if err == nil {
		t.Error("invalid provider key should produce an error")
	}

	_, err = prowl.NewClient(prowl.Config{
		Token:       "1234512345123451234512345123451234512345",
		ProviderKey: "1234123451234512345123451234512345123455",
		Application: stringOfLen(257),
	})
	if err == nil {
		t.Error("invalid application should produce an error")
	}

	_, err = prowl.NewClient(prowl.Config{
		Token:       "1234512345123451234512345123451234512345",
		ProviderKey: "1234123451234512345123451234512345123455",
		Application: "better!",
		APIKeys:     []string{"", "111111111111111111111111111111111111111111"},
	})
	if err == nil {
		t.Error("invalid api keys should produce an error")
	}
}

func ExampleNewClient() {
	// Create a new client e.g. for sending out message.
	client, err := prowl.NewClient(prowl.Config{
		APIKeys:     aValidAPIKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Later you can retrieve the config e.g. to persist and restore
	//it when the program is run again.
	config := client.Config()
	fmt.Println("application: " + config.Application + ".")
	fmt.Println("token: " + config.Token + ".")

	//Now create a new client.
	other, err := prowl.NewClient(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//They should be the same...
	if client.Config().APIKeys[0] == other.Config().APIKeys[0] {
		fmt.Println("both the same.")
	}

	//output:
	//application: prowlgo Example.
	//token: .
	//both the same.
}

func ExampleClient_Add_singleKey() {
	// Create a new client for sending out message.
	client, err := prowl.NewClient(prowl.Config{
		APIKeys:     aValidAPIKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Send something
	remaining, err := client.Add(prowl.PrioNormal, "Test Event", "Test description")
	if err != nil {
		fmt.Println(err)
		return
	}

	if remaining > 0 {
		fmt.Println("some api calls left")
	}

	//output:
	//some api calls left
}

func ExampleClient_Add_multiKey() {
	// Create a new client for sending out message to multiple devices.
	//It's not a lot different from sending to a single device -- it's
	//just multiple api keys in the array this time!
	client, err := prowl.NewClient(prowl.Config{
		APIKeys:     multipleValidAPIKeys,
		Application: "prowlgo Example",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Let's check ...
	//Client should be configured with multiple api keys
	if len(client.Config().APIKeys) != len(multipleValidAPIKeys) {
		fmt.Println("api key count mismatch")
	}

	//Send something
	remaining, err := client.Add(prowl.PrioNormal, "Test Event", "This message is sent to multiple devices...")
	if err != nil {
		fmt.Println(err)
		return
	}

	if remaining > 0 {
		fmt.Println("some api calls left")
	}

	//output:
	//some api calls left
}

func ExampleClient_AddWithURL() {
	// Create a new client for sending out message.
	client, err := prowl.NewClient(prowl.Config{
		APIKeys:     aValidAPIKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Send something with a URL
	remaining, err := client.AddWithURL(prowl.PrioNormal, "Test Event",
		"Test description followed by a URL", "http://github.com/tweithoener/prowlgo", true)
	if err != nil {
		fmt.Println(err)
		return
	}

	if remaining > 0 {
		fmt.Println("api calls left")
	}

	//output:
	//api calls left
}

func TestAdd(t *testing.T) {
	mock.reset()
	defer mock.reset()

	client, err := prowl.NewClient(prowl.Config{
		Application: "prowlgo Example",
	})
	if err != nil {
		t.Error(err)
	}

	if _, err := client.Add(prowl.PrioNormal, "Event", "Description"); err == nil {
		t.Error("add without api key should produce an error")
	}

	client, err = prowl.NewClient(prowl.Config{
		APIKeys:     aValidAPIKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		t.Error(err)
	}
	mock.acceptAPIKeys = false

	if _, err := client.Add(prowl.PrioNormal, "Event", "Description"); err == nil {
		t.Error("illegal api key should produce an error")
	}
	if _, err := client.Add(prowl.PrioNormal, "Event", "Description"); err == nil {
		t.Error("client should be unauthorized now")
	}

	mock.reset()

	client, err = prowl.NewClient(prowl.Config{
		APIKeys:     aValidAPIKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		t.Error(err)
	}

	if _, err := client.Add(-3, "Event", "Description"); err == nil {
		t.Error("invalid prio should produce an error")
	}
	if _, err := client.Add(3, "Event", "Description"); err == nil {
		t.Error("invalid prio should produce an error")
	}
	if _, err := client.Add(prowl.PrioNormal, stringOfLen(1025), "Description"); err == nil {
		t.Error("invalid event should produce an error")
	}
	if _, err := client.Add(prowl.PrioNormal, "Event", stringOfLen(10001)); err == nil {
		t.Error("invalid description should produce an error")
	}
	if _, err := client.AddWithURL(prowl.PrioNormal, "Event", "Description", stringOfLen(257), false); err == nil {
		t.Error("invalid description should produce an error")
	}

	//check if not adding the URL to the description works
	if _, err := client.AddWithURL(prowl.PrioNormal, "Event", "TEXT", "http://URL/", false); err != nil {
		t.Error(err)
	}
	if mock.lastDescription != "TEXT" {
		t.Error("description was altered")
	}

	//check if adding the URL to the description works
	if _, err := client.AddWithURL(prowl.PrioNormal, "Event", "TEXT", "http://URL/", true); err != nil {
		t.Error(err)
	}
	if mock.lastDescription != "TEXT http://URL/" {
		t.Error("description was not composed correctly")
	}

	//description plus url too long for description: should be ok -- URL will be trimmed
	if _, err := client.AddWithURL(prowl.PrioNormal, "Event", stringOfLen(9970), stringOfLen(100), true); err != nil {
		t.Error(err)
	}
	if len(mock.lastDescription) > 10000 {
		t.Error("appending url to description resulted in illegal description")
	}

	//make a succesfull call to get the reset timestamp
	if _, err := client.Add(prowl.PrioNormal, "Event", "Description"); err != nil {
		t.Error(err)
	}

	//Now the call limit is becoming exceeded ....
	mock.callLimit = true

	if _, err := client.Add(prowl.PrioNormal, "Event", "Description"); err == nil {
		t.Error("api call limit reached should produce an error")
	}
	if _, err := client.Add(prowl.PrioNormal, "Event", "Description"); err == nil {
		t.Error("client should be rejected temporarily now")
	}

}

func ExampleClient_Config() {
	// Create a new client e.g. for retrieving an api key
	client, err := prowl.NewClient(prowl.Config{
		ProviderKey: aValidProviderKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Get the config and marshal to JSON
	buf, err := json.Marshal(client.Config())
	if err != nil {
		fmt.Println(err)
		return
	}
	//Write buf to file, shutdown, ..., start, read file,
	config := prowl.Config{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Now create a new client.
	other, err := prowl.NewClient(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//They should be the same...
	if len(client.Config().APIKeys) == len(other.Config().APIKeys) && client.Config().ProviderKey == other.Config().ProviderKey {
		fmt.Println("both the same.")
	}

	//output:
	//both the same.
}

func ExampleClient_Log() {
	toProwlLabel := "--> Prowl"
	// Create a new client for sending out messages
	client, err := prowl.NewClient(prowl.Config{
		APIKeys:      aValidAPIKey,
		Application:  "prowlgo Example",
		ToProwlLabel: &toProwlLabel,
		Logger:       log.New(os.Stdout, "TestLogger: ", 0),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Send out a prowl message that should also be logged by the
	//defined Logger.
	client.Log(prowl.PrioNormal, "Test Event", "Test description which also goes to the log")

	//wait a little to make sure first log message gets written out first.
	<-time.After(1 * time.Second)

	//And the same synn
	client.LogSync(prowl.PrioNormal, "Test Event", "Test description which also goes to the log in sync")

	//output:
	//TestLogger: Test Event: Test description which also goes to the log --> Prowl
	//TestLogger: Test Event: Test description which also goes to the log in sync --> Prowl
}

func TestLogSync(t *testing.T) {
	//this is going to be a little more complex:
	//LogSync()/Log() does not report errors but writes them to the configured log
	//Thus we need a custom logger to check what is in the logs ...
	buf := make([]byte, 1000)
	logbuf := bytes.NewBuffer(buf)
	client, err := prowl.NewClient(prowl.Config{
		APIKeys: aValidAPIKey,
		Logger:  log.New(logbuf, "", 0),
	})
	if err != nil {
		t.Error(err)
	}

	defer mock.reset()

	if !testing.Short() {
		mock.reset()
		//make sure the prowl server mock responds very slow
		mock.wait = 35 * time.Second

		before := time.Now()
		client.Log(prowl.PrioNormal, "TestEvent", "TestDescription")

		//check if call to Log was async (issue #9)
		if before.Add(1 * time.Second).Before(time.Now()) {
			t.Error("function Log() is not async")
		}

		//Log() will timeout after 30 seconds. we should see the error after that.
		<-time.After(31 * time.Second)
		logstr := logbuf.String()
		if !strings.Contains(logstr, "timeout") {
			t.Error("timeout error expected but not found")
		}
	} else {
		t.Log("skipping timeout test in short mode")
	}

	mock.reset()
	mock.acceptAPIKeys = false

	client.LogSync(prowl.PrioNormal, "0123456789012", "01234567890123456789012")
	logstr := logbuf.String()
	if !strings.Contains(logstr, "can't send prowl message") {
		t.Error("send error expected but not found")
	}
	if !strings.Contains(logstr, "(\"0123456...: 01234567890123456...\")") {
		t.Error("shortened message not found in error log")
	}

}

func ExampleClient_Verify_simple() {
	// Create a new client
	//If we just use it to verify an API key we do not need to configure anything.
	client, err := prowl.NewClient(prowl.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	//Verify the provided API key using the client.
	if _, err = client.Verify(singleValidAPIKey); err != nil {
		fmt.Println(err)
	}

	//output:
	//
}

func TestVerify(t *testing.T) {
	defer mock.reset()

	mock.reset()
	//Create a client with provider key
	client, err := prowl.NewClient(prowl.Config{
		ProviderKey: aValidProviderKey,
	})
	if err != nil {
		t.Error(err)
	}

	//Verify a valid api key
	if _, err := client.Verify(singleValidAPIKey); err != nil {
		t.Error(err)
	}

	//And more things that should not work...
	//A provider key is not a api key
	client, err = prowl.NewClient(prowl.Config{})
	if err != nil {
		t.Error(err)
	}

	mock.acceptAPIKeys = false

	//Verify this client -- should produce an error
	if _, err = client.Verify(aValidProviderKey); err == nil {
		t.Error("vrifying an invalid api key should have produced an error")
	}

	mock.reset()
	mock.acceptProviderKey = false

	//A provider key that does not validate is not ap orblem.
	client, err = prowl.NewClient(prowl.Config{
		ProviderKey: singleValidAPIKey,
	})
	if err != nil {
		t.Error(err)
	}
	//Verify this key -- will work out. The provider key is not verified
	//but you wont profit from a higher api call limit either (which your
	//valid provider key might be white listed for)
	if _, err = client.Verify(singleValidAPIKey); err != nil {
		t.Error(err)
	}

	//And finally no key at all -- can't work either
	client, err = prowl.NewClient(prowl.Config{})
	if err != nil {
		t.Error(err)
	}
	//Verify an empty key -- should produce an error.
	//Actually that shouldn't even go out to the Prowl server.
	//No api call spent on that.
	if _, err := client.Verify(""); err == nil {
		t.Error("verifying an empty kes should have produced an error")
	}

	mock.reset()
	mock.incomplete = true

	if _, err := client.Verify(singleValidAPIKey); err == nil {
		t.Error("incomplete response should have produced an error")
	}

	mock.stop()

	if _, err := client.Verify(singleValidAPIKey); err == nil {
		t.Error("server not responding should have produced an error")
	}
}

func ExampleClient_RetrieveToken() {
	//A client that is good to retrieve a token
	client, err := prowl.NewClient(prowl.Config{
		ProviderKey: aValidProviderKey,
	})
	if err != nil {
		fmt.Println(err)
	}

	//Let's retrieve the token and the approve URL
	//which is in the first return value which we ignore here
	//as there is little we can do with it in this example)
	approveURL, err := client.RetrieveToken()
	if err != nil {
		fmt.Println(err)
	}

	//Now present the approveURL to the user and wait for his approval.
	fmt.Println("Please approve api key request at:", approveURL)

	//... user approves ...

	//User has approved the URL, let's continue
	_, err = client.RetrieveAPIKey()
	if err != nil {
		fmt.Println(err)
	}

	//output:
	//Please approve api key request at: https://www.prowlapp.com/retrieve.php?token=c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7
}

func TestRetrieveTokenAndAPIKey(t *testing.T) {
	mock.reset()
	defer mock.reset()

	client, err := prowl.NewClient(prowl.Config{})
	if err != nil {
		t.Error(err)
	}

	//retrieve token without provider key will fail
	if _, err := client.RetrieveToken(); err == nil {
		t.Error("retrieve token without provider key should produce an error")
	}

	//retrieve api key with invalid provider key/token will fail
	if _, err := client.RetrieveAPIKey(); err == nil {
		t.Error("retrieve api key with invalid provider key/token should produce an error")
	}

	client, err = prowl.NewClient(prowl.Config{
		Token: "0987609876098760987609876098760987609876",
	})
	if err != nil {
		t.Error(err)
	}

	//retrieve api key with invalid provider key will fail
	if _, err := client.RetrieveAPIKey(); err == nil {
		t.Error("retrieve api key with invalid provider key should produce an error")
	}

	client, err = prowl.NewClient(prowl.Config{
		ProviderKey: "0123401234012340123401234012340123401234",
		Token:       "0987609876098760987609876098760987609876",
	})
	if err != nil {
		t.Error(err)
	}

	mock.acceptProviderKey = false
	//retrieve token with invalid provider key will fail
	if _, err := client.RetrieveToken(); err == nil {
		t.Error("retrieve token with invalid provider key should produce an error")
	}

	//retrieve api key with invalid provider key will fail
	if _, err := client.RetrieveAPIKey(); err == nil {
		t.Error("retrieve api key wit invalid provider key should produce an error")
	}

	mock.acceptProviderKey = true
	mock.acceptToken = false

	//retrieve api key with invalid token will fail
	if _, err := client.RetrieveAPIKey(); err == nil {
		t.Error("retrieve api key with invalid token should produce an error")
	}

	mock.stop()

	//both request against stopped server will fail
	if _, err := client.RetrieveAPIKey(); err == nil {
		t.Error("retrieve api key against server not running should fail")
	}
	if _, err := client.RetrieveAPIKey(); err == nil {
		t.Error("retrieve api key against server not running should fail")
	}

	mock.reset()

	//let's go through the process again with no error:
	if _, err := client.RetrieveToken(); err != nil {
		t.Error(err)
	}
	if _, err := client.RetrieveAPIKey(); err != nil {
		t.Error(err)
	}

	if client.Config().Token != "c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7" {
		t.Error("wrong token in client config")
	}
	found := false
	for _, key := range client.Config().APIKeys {
		if key == "3fa013fa013fa013fa013fa013fa013fa013fa01" {
			found = true
			break
		}
	}
	if !found {
		t.Error("new api key was not found in client config")
	}

	mock.incomplete = true

	//both should produce an error
	if _, err := client.RetrieveToken(); err == nil {
		t.Error("incomplete response should produce an error")
	}
	if _, err := client.RetrieveAPIKey(); err == nil {
		t.Error("incomplete response should produce an error")
	}

}

func TestReset(t *testing.T) {
	mock.reset()
	defer mock.reset()

	resetTS := time.Now().Add(2 * time.Minute).Unix()
	mock.resetTS = resetTS

	client, err := prowl.NewClient(prowl.Config{})
	if err != nil {
		t.Error(err)
	}

	//Let's make a request. In the answer the client will find the reset timestamp which is
	//then stored in the client instance.
	if _, err := client.Verify("0123401234012340123401234012340123401234"); err != nil {
		t.Error(err)
	}

	if client.Reset().Unix() != resetTS {
		t.Error("timestamp does not match")
	}

}

func ExampleClient_AddAPIKey() {

	//First create a new client, with a single API key
	client, err := prowl.NewClient(prowl.Config{
		APIKeys: []string{"5a7d185a7d185a7d185a7d185a7d185a7d185a7d"},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Now add another API key
	if err := client.AddAPIKey("55df3a55df3a55df3a55df3a55df3a55df3a55df"); err != nil {
		fmt.Println(err)
		return
	}

	//And add the same key again. It will not produce an error but the key will not be added.
	if err := client.AddAPIKey("55df3a55df3a55df3a55df3a55df3a55df3a55df"); err != nil {
		fmt.Println(err)
		return
	}

	//We should now have two API keys in the configuratoin
	if len(client.Config().APIKeys) == 2 {
		fmt.Println("two api keys in config")
	}

	//Now let's remove the first API key
	if err := client.RemoveAPIKey("5a7d185a7d185a7d185a7d185a7d185a7d185a7d"); err != nil {
		fmt.Println(err)
		return
	}

	//We should now have one API key in the configuratoin
	if len(client.Config().APIKeys) == 1 {
		fmt.Println(client.Config().APIKeys[0])
	}

	//output:
	//two api keys in config
	//55df3a55df3a55df3a55df3a55df3a55df3a55df
}

func TestAddRemoveAPIKeys(t *testing.T) {

	client, err := prowl.NewClient(prowl.Config{})
	if err != nil {
		t.Error(err)
	}

	//Check illegal API keys
	if err := client.AddAPIKey(""); err == nil {
		t.Error("illegal API key should produce an error")
	}

	if err := client.AddAPIKey(stringOfLen(41)); err == nil {
		t.Error("illegal API key should produce an error")
	}
	if err := client.RemoveAPIKey(""); err == nil {
		t.Error("illegal API key should produce an error")
	}
	if err := client.RemoveAPIKey(stringOfLen(41)); err == nil {
		t.Error("illegal API key should produce an error")
	}

	//Removing something that is not there shouldn't produce an error
	if err := client.RemoveAPIKey("0123401234012340123401234012340123401234"); err != nil {
		t.Error("error removing API key that was not added before")
	}

	//We expect an empty array of API keys at this point
	if len(client.Config().APIKeys) != 0 {
		t.Error("client config not correct")
	}

	//Add two API keys, then send a message and check that the "apikey" parameter of
	//the http request was correct.
	key1 := "1111111111111111111111111111111111111111"
	key2 := "2222222222222222222222222222222222222222"

	if err := client.AddAPIKey(key1); err != nil {
		t.Error(err)
	}
	if err := client.AddAPIKey(key2); err != nil {
		t.Error(err)
	}
	if _, err := client.Add(prowl.PrioNormal, "TestEvent", "TestDescription"); err != nil {
		t.Error(err)
	}

	p1 := key1 + "," + key2
	p2 := key2 + "," + key1
	if mock.lastAPIKey != p1 && mock.lastAPIKey != p2 {
		log.Println(mock.lastAPIKey)
		t.Error("apikey request parameter is not correct")
	}
}

// ----------------------------------------------------------------------------------------------
// Mocking a https server during testing

func prowlMockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/xml")

	<-time.After(mock.wait)

	if mock.internalError {
		w.WriteHeader(500)
		fmt.Sprintln(w, internalError)
		return
	}

	if mock.incomplete {
		w.WriteHeader(200)
		fmt.Sprintln(w, incompleteXML)
		return
	}

	remaining := 992
	if mock.callLimit {
		remaining = 0
	}

	switch r.URL.Path {
	case "/publicapi/add":
		if r.FormValue("apikey") == "" {
			w.WriteHeader(401)
			fmt.Fprintf(w, apiKeyRequired)
			return
		}
		if remaining == 0 {
			w.WriteHeader(406)
			fmt.Fprintf(w, callLimitExceeded)
			return
		}
		if !mock.acceptAPIKeys {
			w.WriteHeader(401)
			fmt.Fprintf(w, invalidAPIKey)
			return
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, add200, remaining, mock.resetTS)
		mock.lastDescription = r.FormValue("description")
		mock.lastAPIKey = r.FormValue("apikey")

	case "/publicapi/verify":
		if r.URL.Query().Get("apikey") == "" {
			w.WriteHeader(401)
			fmt.Fprintf(w, apiKeyRequired)
			return
		}

		if mock.acceptAPIKeys {
			w.WriteHeader(200)
			fmt.Fprintf(w, verify200, remaining, mock.resetTS)
			return
		}

		w.WriteHeader(401)
		fmt.Fprintf(w, invalidAPIKey)

	case "/publicapi/retrieve/token":
		if r.URL.Query().Get("providerkey") == "" || !mock.acceptProviderKey {
			w.WriteHeader(401)
			fmt.Fprintf(w, providerKeyRequired)
			return
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, retrieveToken200, remaining, mock.resetTS)

	case "/publicapi/retrieve/apikey":
		if r.URL.Query().Get("token") == "" {
			w.WriteHeader(401)
			fmt.Fprintf(w, tokenIsRequired)
			return
		}
		if r.URL.Query().Get("providerkey") == "" || !mock.acceptProviderKey {
			w.WriteHeader(401)
			fmt.Fprintf(w, providerKeyRequired)
			return
		}

		if mock.acceptToken {
			w.WriteHeader(200)
			fmt.Fprintf(w, retrieveAPIKey200, remaining, mock.resetTS)
			return
		}

		w.WriteHeader(409)
		fmt.Fprintf(w, tokenNotApproved)
	}
}

type mockServer struct {
	acceptToken       bool
	acceptProviderKey bool
	acceptAPIKeys     bool
	callLimit         bool
	incomplete        bool
	internalError     bool
	wait              time.Duration
	resetTS           int64
	lastDescription   string
	lastAPIKey        string
	server            *httptest.Server
}

func (ms *mockServer) start() {
	if ms.server != nil {
		return
	}
	ms.server = httptest.NewServer(http.HandlerFunc(prowlMockHandler))

	mockURL, err := url.Parse(mock.server.URL)
	if err != nil {
		log.Fatal("error parsing mock server url:", err)
	}

	//create a http client using that transport
	testClient := &http.Client{
		Transport: RewriteTransport{
			Transport: &http.Transport{
				Dial: func(network, addr string) (net.Conn, error) {
					return net.DialTimeout(network, addr, 3*time.Second)
				},
			},
			URL: mockURL,
		},
	}

	//inject this client
	http.DefaultClient = testClient

}

func (ms *mockServer) stop() {
	if ms.server == nil {
		return
	}
	ms.server.Close()
	ms.server = nil
}

func (ms *mockServer) reset() {
	ms.acceptAPIKeys = true
	ms.acceptProviderKey = true
	ms.acceptToken = true
	ms.incomplete = false
	ms.internalError = false
	ms.callLimit = false
	ms.wait = 0
	ms.start()
	ms.resetTS = time.Now().Add(37 * time.Minute).Unix()
}

var mock = &mockServer{}

func TestMain(m *testing.M) {
	//setup a mock http server
	mock.reset()

	//run the tests
	os.Exit(m.Run())
}

func stringOfLen(len int) string {
	ret := ""
	for i := 0; i < len; i++ {
		ret = ret + "x"
	}
	return ret
}

var (
	singleValidAPIKey    = "e192384beae856efa6dda87d6a00837cf968bd8c"
	aValidAPIKey         = []string{"e192384beae856efa6dda87d6a00837cf968bd8c"}
	multipleValidAPIKeys = []string{"e192384beae856efa6dda87d6a00837cf968bd8c", "e19238423ae856efa6ddadf34a00837cf968bd8c", "e17ef34beae856efa6dda87d6a0082130168bd8c"}
	aValidProviderKey    = "0267157cc27a27f99ad23d1f785f0e7897df0d6b"
)

const invalidAPIKey = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<error code="401">Invalid API key</error>
</prowl>
`

const apiKeyRequired = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<error code="401">API key is required</error>
</prowl>
`

const verify200 = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<success code="200" remaining="%d" resetdate="%d" />
</prowl>
`

const add200 = verify200

const providerKeyRequired = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<error code="401">Provider key is required.</error>
</prowl>
`

const callLimitExceeded = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<error code="406">Not acceptable, your IP address has exceeded the API limit</error>
</prowl>
`

const invalidProviderKey = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<error code="401">Invalid provider key.</error>
</prowl>
`

const retrieveToken200 = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<success code="200" remaining="%d" resetdate="%d" />
<retrieve token="c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7" url="https://www.prowlapp.com/retrieve.php?token=c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7c3fb7" />
</prowl>
`

const tokenNotApproved = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<error code="409">The user has not approved your access.</error>
</prowl>
`

const tokenIsRequired = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<error code="400">Token is required.</error>
</prowl>
`

const retrieveAPIKey200 = `
<?xml version="1.0" encoding="UTF-8"?>
<prowl>
<success code="200" remaining="%d" resetdate="%d" />
<retrieve apikey="3fa013fa013fa013fa013fa013fa013fa013fa01" />
</prowl>
`

const internalError = `
<prowl>
<error code="500">Somethign went wrong.</error>
</prowl>
`

const incompleteXML = `
<prowl>
<error code="500">Somethign we
`

type RewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (rwt RewriteTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req.URL.Scheme = rwt.URL.Scheme
	req.URL.Host = rwt.URL.Host
	req.URL.Path = path.Join(rwt.URL.Path, req.URL.Path)

	rt := rwt.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}

	resp, err = rt.RoundTrip(req)
	return
}
