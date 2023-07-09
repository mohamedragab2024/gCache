package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve the value for a given key from the distributed key-value store.",
	Long: `The get command retrieves the value associated with the specified key
			from the in-memory distributed key-value store. It takes the desired key 
			as an argument and sends a get command to the server, 
			returning the corresponding value from the distributed store.`,
	Run:  getExecute,
	Args: cobra.ExactArgs(1),
}

func getExecute(cmd *cobra.Command, args []string) {
	key := args[0]
	val, err := gClient.Get(context.Background(), []byte(key))
	if err != nil {
		fmt.Printf("failed to get [%s] [%v] \n", key, err)
		return
	}
	fmt.Println(string(val))
}
