package cmd

import (
	"bufio"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/ragoob/gCache/pkg/grpc/client"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var serverAddr string
var gClient *client.Client

var rootCmd = &cobra.Command{
	Use:   "gcache",
	Short: "gcache is a simple go in-memory Distributed key-value store",
	Long: `root command serves as the entry point for the interactive CLI client
		of the in-memory distributed key-value store. It provides a command-line interface
		for interacting with the key-value store server. A Fast and Flexible in-memory 
		Distributed key-value store built with love in Go.`,
	Run: runRoot,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		gClient, err = client.Connect(serverAddr, client.Options{})
		if err != nil {
			return err
		}
		fmt.Printf("Connected to the server with port %s \n", serverAddr)
		return nil
	},
}

func Execute() {
	rootCmd.Flags().StringVarP(&serverAddr, "serveraddr", "s", ":8080", "Server address")
	defer gClient.Close()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	scanner := bufio.NewScanner(os.Stdin)
	welcomePrint()

	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if len(input) == 0 {
			continue
		}

		handleCommand(cmd, input)

	}
}

func handleCommand(cmd *cobra.Command, input string) {
	// Parse the command and arguments
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}
	command := parts[0]
	arguments := parts[1:]

	switch command {
	case "SET", "set":
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Error SET command: %v\n", r)
			}
		}()
		setCmd.Run(cmd, arguments)
	case "GET", "get":
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Error GET command: %v\n", r)
			}
		}()
		getCmd.Run(cmd, arguments)
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}

}

func welcomePrint() {
	data := [][]string{
		{"SET", "foo", "bar value"},
		{"set", "foo", "bar value"},
		{"GET", "foo"},
		{"get", "foo"},
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
