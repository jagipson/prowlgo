package prowlgo

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func ExampleNewClient() {
	// Create a new client e.g. for sending out message.
	client, err := NewClient(Config{
		APIKey:      aValidAPIKey,
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
	other, err := NewClient(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//They should be the same...
	if client.Config().APIKey == other.Config().APIKey {
		fmt.Println("both the same.")
	}

	//output:
	//application: prowlgo Example.
	//token: .
	//both the same.
}

func ExampleClient_Add() {
	// Create a new client for sending out message.
	client, err := NewClient(Config{
		APIKey:      aValidAPIKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Send something
	remaining, err := client.Add(PrioNormal, "Test Event", "Test description")
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
	client, err := NewClient(Config{
		APIKey:      aValidAPIKey,
		Application: "prowlgo Example",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Send something with a URL
	remaining, err := client.AddWithURL(PrioNormal, "Test Event",
		"Test description followed by a URL", "http://github.com/tweithoener/prowlgo")
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

func ExampleClient_Config() {
	// Create a new client e.g. for retrieving an api key
	client, err := NewClient(Config{
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
	config := Config{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Now create a new client.
	other, err := NewClient(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//They should be the same...
	if client.Config().APIKey == other.Config().APIKey && client.Config().ProviderKey == other.Config().ProviderKey {
		fmt.Println("both the same.")
	}

	//output:
	//both the same.
}

func ExampleClient_Log() {
	toProwlLabel := "--> Prowl"
	// Create a new client for sending out messages
	client, err := NewClient(Config{
		APIKey:       aValidAPIKey,
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
	client.Log(PrioNormal, "Test Event", "Test description which also goes to the log")

	//output:
	//TestLogger: Test Event: Test description which also goes to the log --> Prowl
}

func ExampleClient_Verify_simple() {
	// Create a new working client e.g. for sending out messages
	client, err := NewClient(Config{
		APIKey: aValidAPIKey,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//Verify this client
	if _, err = client.Verify(); err != nil {
		fmt.Println(err)
	}

	//output:
	//
}

func ExampleClient_Verify_more() {
	//Create a client with api key and provider key
	//Might be used to send messages and thereby profiting
	//from a raised api limit for the provider key
	client, err := NewClient(Config{
		APIKey:      aValidAPIKey,
		ProviderKey: aValidProviderKey,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//Verify this client -- should also work out
	if _, err := client.Verify(); err != nil {
		fmt.Println(err)
	}

	//Next a client with only a provider key
	//That won't verify.
	client, err = NewClient(Config{
		ProviderKey: aValidProviderKey,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//Verify this client -- should not work
	if _, err = client.Verify(); err != nil {
		fmt.Println("intentional error")
	}

	//And more things that should not work...
	//A api key is not a provider key
	client, err = NewClient(Config{
		ProviderKey: aValidAPIKey,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//Verify this client -- should produce an error
	if _, err = client.Verify(); err != nil {
		fmt.Println("intentional error")
	}

	//Also wrong
	client, err = NewClient(Config{
		APIKey:      aValidProviderKey,
		ProviderKey: aValidAPIKey,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//Verify this client -- should produce an error
	if _, err = client.Verify(); err != nil {
		fmt.Println("intentional error")
	}

	//And finally no key at all -- can't work either
	client, err = NewClient(Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	//Verify this client -- should produce an error.
	//Actually that shouldn't even go out to the Prowl server.
	//No api call spent on that.
	if _, err := client.Verify(); err != nil {
		fmt.Println("intentional error")
	}

	//output:
	//intentional error
	//intentional error
	//intentional error
	//intentional error
}

func ExampleClient_RetrieveToken() {
	//A client that is good to retrieve a token
	client, err := NewClient(Config{
		ProviderKey: aValidProviderKey,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//Let's retrieve the token and the approve URL
	//which is in the first return value which we ignore here
	//as there is little we can do with it in this example)
	_, err = client.RetrieveToken()
	if err != nil {
		fmt.Println(err)
	}

	//Now present the approveURL to the user and wait for his approval.
	//You can then call Client.RetrieveAPIKey which will fail now as
	//we are not waiting...
	_, err = client.RetrieveAPIKey()
	if err != nil {
		fmt.Println("intentional error")
	}

	//output:
	//intentional error
}
