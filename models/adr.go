package models

import "errors"

type Adr struct {
	Title  string
	Author string
	Date   string
	Status string
}

type History struct {
	Date   string
	Status string
	Memo   string
}

var statusMapper = map[string]bool{
	"proposed":   true,
	"accepted":   true,
	"rejected":   true,
	"deprecated": true,
	"suspended":  true,
}

var StatusHeader = "Status"

const A = "Status"

var HistoryHeader = "History"

func CheckStatus(status string) error {
	if _, ok := statusMapper[status]; ok {
		return nil
	}
	return errors.New("invalid status used")
}
