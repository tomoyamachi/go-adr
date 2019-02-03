package cmd

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/tomoyamachi/go-adr/models"

	"github.com/spf13/viper"

	"github.com/tomoyamachi/go-adr/renderer"

	"github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
)

var changeCmd = &cobra.Command{
	Use:   "change",
	Short: "change status",
	Long:  `change status`,
	RunE:  changeStatus,
}

func init() {
	RootCmd.AddCommand(changeCmd)
	changeCmd.PersistentFlags().String(
		"new-status", "",
		"set new status (proposed, accepted, rejected, deprecated, superseded, etc.",
	)
	viper.BindPFlag("new-status", changeCmd.PersistentFlags().Lookup("new-status"))

	changeCmd.PersistentFlags().String(
		"memo", "",
		"add history why you changed status",
	)
	viper.BindPFlag("memo", changeCmd.PersistentFlags().Lookup("memo"))
	viper.SetDefault("memo", "")
}

func changeStatus(cmd *cobra.Command, args []string) (err error) {
	status := viper.GetString("new-status")
	if err = models.CheckStatus(status); err != nil {
		return err
	}

	filename := args[0]
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log15.Error("Failed to read file", filename, err)
		return err
	}

	addHistory := models.History{
		Date:   time.Now().Format("2006-01-02"),
		Status: status,
		Memo:   viper.GetString("memo"),
	}

	md := renderer.StatusChange(fileBytes, &addHistory)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(md)
	if err != nil {
		return err
	}

	log15.Info("Update target file", "file", filename)
	return nil
}
