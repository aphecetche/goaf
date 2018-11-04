package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var stageCmd = &cobra.Command{
	Use:   "stage",
	Short: "Handle staging",
	Run:   stage,
}

func stage(cmd *cobra.Command, args []string) {
	fmt.Println("here I am")
	servers := viper.GetStringSlice("servers")
	for _, s := range servers {
		fmt.Println(s)
	}
}

var request string
var filter string

func init() {
	RootCmd.AddCommand(stageCmd)
	stageCmd.PersistentFlags().StringVarP(&request, "request", "r", "", "filelist of the files to be requestd")
	stageCmd.PersistentFlags().StringVarP(&filter, "filter", "f", "", "filter to be used")
}
