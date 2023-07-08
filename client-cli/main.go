package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/ragoob/gCache/client"
)

func main() {
	var (
		serveraddr = flag.String("serveraddr", "", "listen address of the server")
		command    = flag.String("command", "", "Cache command get , set")
		key        = flag.String("key", "", "Cache key")
		value      = flag.String("val", "", "Cache value for wirte")
	)
	flag.Parse()

	c, err := client.Connect(*serveraddr, client.Options{})
	if err != nil {
		panic(err.Error())
	}

	if *command == "SET" {
		c.Set(context.Background(), []byte(*key), []byte(*value), 5)
		fmt.Println("Successfully Set key")
	} else if *command == "GET" {
		val, err := c.Get(context.Background(), []byte(*key))
		if err != nil {
			panic(err)
		}

		fmt.Println(string(val))
	} else {
		panic("Invalid command")
	}

	c.Close()

}
