package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/ragoob/gCache/client"
)

func main() {
	var (
		serveraddr = flag.String("serveraddr", "", "listen address of the server")
	)
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	c, err := client.Connect(*serveraddr, client.Options{})
	if err != nil {
		panic(err.Error())
	}

	welcomePrint()
	for {
		fmt.Print(">> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSuffix(input, "\n")
		if strings.TrimSpace(input) == "exit" {
			break
		}
		handleCommand(input, c)
	}

	c.Close()

}

func handleCommand(input string, c *client.Client) {
	cmd := parseCommand(input)
	if len(cmd) < 2 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s SET KEY VALUE \n GET KEY .. \n",
			os.Args[0])
		os.Exit(1)
	}
	if cmd[0] == "SET" {
		if len(cmd) < 3 {
			fmt.Fprintf(os.Stderr,
				"Usage: %s SET KEY VALUE \n",
				os.Args[0])
			os.Exit(1)
		}
		err := c.Set(context.Background(), []byte(cmd[1]), []byte(cmd[2]), 5)
		if err != nil {
			fmt.Printf("failed to set [%s] [%v]", cmd[1], err)
			return
		}
		fmt.Println("Successfully Set key")
	} else if cmd[0] == "GET" {
		val, err := c.Get(context.Background(), []byte(cmd[1]))
		if err != nil {
			fmt.Printf("failed to get [%s] [%v]", cmd[1], err)
			return
		}

		fmt.Println(string(val))
	} else {
		fmt.Fprintf(os.Stderr,
			"Usage: %s SET KEY VALUE \n GET KEY .. \n",
			os.Args[0])
		os.Exit(1)
	}
}

func parseCommand(input string) []string {
	split := strings.Split(input, " ")

	var result []string
	for _, s := range split {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func welcomePrint() {
	data := [][]string{
		{"SET", "foo", "foo value"},
		{"GET", "foo"},
	}

	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Command", "KEY", "VALUE"})

	for _, row := range data {
		table.Append(row)
	}
	table.SetBorder(false)
	table.SetColumnSeparator("│")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}
