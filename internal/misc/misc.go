package misc

import (
	"fmt"
	"strings"
	"time"
)

func DateFromFileName(fn string) string {
	parts := strings.Split(fn, "-")
	t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2]))
	if err != nil {
		return ""
	}
	return t.Format("01/02/2006")
}
