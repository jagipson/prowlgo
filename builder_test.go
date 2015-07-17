package prowlgo

import (
	"fmt"
	"log"
	"os"
)

func ExampleBuilder_simple() {
	//To create a prowl client which is goo for sending out
	//message we will need an api key and the application string won't hurt
	client, err := NewBuilder().
		SetAPIKey(aValidAPIKey).
		SetApplication("prowlgo Test").
		Build()
	if err != nil {
		fmt.Println(err)
	}

	//Let's see if we got what we expected...
	if client.Config().APIKey == aValidAPIKey {
		fmt.Println("api key is ok")
	}
	if client.Config().Application == "prowlgo Test" {
		fmt.Println("application is ok")
	}
	if len(client.Config().ProviderKey) != 0 {
		fmt.Println("error " + client.Config().ProviderKey + ".")
	}
	if len(client.Config().Token) != 0 {
		fmt.Println("error " + client.Config().Token + ".")
	}
	if *client.Config().ToProwlLabel != defaultToProwlLabel {
		fmt.Println("error " + *client.Config().ToProwlLabel + ".")
	}

	//output:
	//api key is ok
	//application is ok
}

func ExampleBuilder_more() {
	toProwlLabel := "--> prowl"

	//To create a prowl client which is goo for sending out
	//message we will need an api key and the application string won't hurt
	client, err := NewBuilder().
		SetAPIKey(aValidAPIKey).
		SetProviderKey(aValidProviderKey).
		SetToken("0123456789012345678901234567890123456789").
		SetApplication("prowlgo Test").
		SetToProwlLabel(toProwlLabel).
		SetLogger(log.New(os.Stdout, "", 0)).
		Build()
	if err != nil {
		fmt.Println(err)
	}

	//Let's see if we got what we expected...
	if client.Config().APIKey == aValidAPIKey {
		fmt.Println("api key is ok")
	}
	if client.Config().Application == "prowlgo Test" {
		fmt.Println("application is ok")
	}
	if client.Config().ProviderKey == aValidProviderKey {
		fmt.Println("provider key is ok")
	}
	if client.Config().Token == "0123456789012345678901234567890123456789" {
		fmt.Println("token is ok")
	}
	if *client.Config().ToProwlLabel == toProwlLabel {
		fmt.Println("toProwlLabel is ok")
	}

	//output:
	//api key is ok
	//application is ok
	//provider key is ok
	//token is ok
	//toProwlLabel is ok
}
