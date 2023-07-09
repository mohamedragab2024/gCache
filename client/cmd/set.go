package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a key-value pair in the distributed key-value store.",
	Long: `The set command allows you to set a key-value pair in the in-memory 
			distributed key-value store. It takes the specified key and value as arguments 
			and sends the set command to the server, storing the provided value for 
			the given key in the distributed store.`,
	Run:  setExecute,
	Args: cobra.ExactArgs(2),
}

func setExecute(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]
	err := gClient.Set(context.Background(), []byte(key), []byte(value), 5)
	if err != nil {
		fmt.Printf("failed to set [%s] [%v] \n", key, err)
		return
	}
	fmt.Println("Successfully Set key")
}
