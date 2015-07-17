# prowlgo
golang Interface to Prowl: Send Push Notifications to iOS Devices From Your go Application

## Quick-Start

 1. Go to http://www.prowlapp.com/ and crea)te yourself an API key and a provider key.
 1. Install the package

		go get github.com/tweithoener/prowlgo

 1. Get a minimal example going. Grab the code from below, put it into a file (say `prowling.go`), replace the dummy API key with a real one, and run the program using `go run prowling.go`

		package main

		import (
			"fmt"
			prowl "github.com/tweithoener/prowlgo"
		)

		func main() {
			//Create the client.
			client, err := prowl.NewClient(prowl.Config{
				APIKey:      "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde", //Replace with something valid!
				Application: "prowlgo Demo",
			})
			if err != nil {
				fmt.Println("Can't create prowl client: " + err.Error())
				return
			}

			//And send the message.
			remaining, err := client.Add(prowl.PrioNormal, "Hello World", "Your first message via prowlgo")
			if err != nil {
				fmt.Println("can't send message: " + err.Error())
				return
			}

			fmt.Printf("remaining api calls: %s\n", remaining)
		}

 1. Congratulations your first prowl message was just delivered to your device.
 1. Make sure everything is working by running the tests. The tests require a API key and a provider key which you already created in the first step. Now copy the file prowlgo/setup_to_test.go.sample to setup_the_test.go

		cd prowlgo
		cp setup_the_test.go.sample setup_the_test.go

 1. open setup_the_test.go with an editor and replace the dummy keys with the keys you created at prowlapp.com
 1. save prowlgo/setup_the_test.go
 1. thenrun `go test` in the prowl directory. All tests should be passed.

## Documentation

The package is documented using godoc. Thre resulting documentation can be found here: http://godoc.org/github.com/tweithoener/prowlgo
Check it out it is full of code examples. Start at type Client and browse through it. It's easy to understand.

## TODO

Looks like I forgot that you can feed multiple API keys into one request to the prowl server. I'll take care of this... very soon.
Besides that, let me know about any problems that you detect.

## What's about that other golang prowl package

I must have been really blind. I checked the third party libraries list at http://www.prowlapp.com/api.php and didn't realize there already is another golang project. It's also on github: https://github.com/rem7/goprowl . When I finally realized I wasn't the first, I was almost done and decided to publish mine too. So now we have prowlgo and goprowl. Happy confusion!

