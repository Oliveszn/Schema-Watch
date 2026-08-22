package alert

import (
	"fmt"
	"io"
	"os"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBold   = "\033[1m"
)

// Alerter is anything that can be notified when a schema diff occurs
type Alerter interface {
	Alert(diff *schema.Diff)
}

type Console struct {
	out   io.Writer
	color bool
}

func NewConsole() *Console {
	return &Console{out: os.Stdout, color: true}
}

func NewConsoleWithWriter(w io.Writer, color bool) *Console {
	return &Console{out: w, color: color}
}

func (c *Console) Alert(diff *schema.Diff) {
	if diff == nil {
		return
	}

	label := "CHANGE"
	headerColor := colorYellow
	if diff.Breaking {
		label = "BREAKING CHANGE"
		headerColor = colorRed
	}

	fmt.Fprintf(c.out, "\n%s[schema-watch] %s on %s%s\n", c.wrap(colorBold+headerColor), label, diff.Endpoint, c.wrap(colorReset))

	for _, change := range diff.Changes {
		switch change.Type {
		case schema.FieldAdded:
			fmt.Fprintf(c.out, "%s  + %s added (%s)%s\n", c.wrap(colorGreen), change.Path, change.NewType, c.wrap(colorReset))
		case schema.FieldRemoved:
			fmt.Fprintf(c.out, "%s  - %s removed (was %s)%s\n", c.wrap(colorRed), change.Path, change.OldType, c.wrap(colorReset))
		case schema.FieldTypeChanged:
			fmt.Fprintf(c.out, "%s  ~ %s changed type: %s -> %s%s\n", c.wrap(colorYellow), change.Path, change.OldType, change.NewType, c.wrap(colorReset))
		}
	}
}

func (c *Console) wrap(code string) string {
	if !c.color {
		return ""
	}
	return code
}
