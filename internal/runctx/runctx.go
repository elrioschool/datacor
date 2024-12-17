package runctx

import (
	"fmt"
	"strings"
	"time"
)

type RunContext struct {
	NewTxnReport  string
	PrevTxnReport string
}

func (ctx *RunContext) GetNewReportDate() string {
	parts := strings.Split(ctx.NewTxnReport, "-")
	t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2]))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return t.Format("01/02/2006")
}
