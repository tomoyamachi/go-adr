package cmd

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tomoyamachi/go-adr/models"
	//"github.com/tomoyamachi/go-adr/models"
)

// outputCmd represents output golang file for future-architect/vuls
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create ADR template",
	Long:  `Create ADR template`,
	RunE:  createAdr,
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().String(
		"status", "proposed",
		"set ADR status (proposed, accepted, rejected, deprecated, superseded, etc.",
	)
	viper.BindPFlag("status", createCmd.PersistentFlags().Lookup("status"))
	viper.SetDefault("status", "proposed")
}

func createAdr(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("%v", args)
	status := viper.GetString("status")
	filename := strings.Join(args, "_")
	user, err := user.Current()
	if err != nil {
		return err
	}
	adr := models.Adr{
		Title:  strings.Join(args, " "),
		Date:   time.Now().Format("2006-01-02"),
		Author: user.Username,
		Status: status,
	}
	tmpl, err := template.New("adr").Parse(templateAdr)
	if err != nil {
		return err
	}
	outputFile := time.Now().Format("20060102150405") + "_" + filename + ".md"
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl.Execute(f, adr)
	return nil
}

const templateAdr = `
# {{.Title}}

{{.Author}}

## Status

{{.Status}}

## History

| Date | Status | Memo |
|--|--|--|
| {{.Date}} | {{.Status}} | create this file |

## Context

> what is the issue that we're seeing that is motivating this decision or change.

## Decision

> what is the change that we're actually proposing or doing.

## Consequences

> what becomes easier or more difficult to do because of this change.
`
