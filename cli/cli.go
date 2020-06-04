package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fatih/color"
	"github.com/rmrf/robo/config"
)

// Template helpers.
var helpers = template.FuncMap{
	"magenta": color.MagentaString,
	"yellow":  color.YellowString,
	"green":   color.GreenString,
	"black":   color.BlackString,
	"white":   color.WhiteString,
	"blue":    color.BlueString,
	"cyan":    color.CyanString,
	"red":     color.RedString,
}

// List template.
var list = `
{{range .Tasks}}  {{cyan .Name}} – {{.Summary}}
{{end}}
`

// Variables template.
var variables = `
{{- range $parent, $v := .Variables}}
{{- range $k, $v := $v }}
    {{cyan "%s.%s" $parent $k }}: {{$v}}
{{- end}}
{{end}}
`

// Help template.
var help = `
  {{cyan "Usage:"}}

    {{.Name}} {{.Usage}}

  {{cyan "Description:"}}

    {{.Summary}}
{{with .Examples}}
  {{cyan "Examples:"}}
  {{range .}}
    {{.Description}}
    $ {{.Command}}
  {{end}}{{end}}
`

// ListVariables outputs the variables defined.
func ListVariables(c *config.Config) {
	tmpl := t(variables)

	if c.Templates.Variables != "" {
		tmpl = t(c.Templates.Variables)
	}

	tmpl.Execute(os.Stdout, c)
}

// List outputs the tasks defined.
func List(c *config.Config) {
	tmpl := t(list)

	if c.Templates.List != "" {
		tmpl = t(c.Templates.List)
	}

	tmpl.Execute(os.Stdout, c)
}

// Help outputs the task help.
func Help(c *config.Config, name string) {
	task, ok := c.Tasks[name]

	if !ok {
		Errorf("undefined task %q", name)
	}

	tmpl := t(help)

	if c.Templates.Help != "" {
		tmpl = t(c.Templates.Help)
	}

	tmpl.Execute(os.Stdout, task)
}

// Run the task.
func Run(c *config.Config, name string, args []string) error {
	task, ok := c.Tasks[name]
	if !ok {
		Errorf("undefined task %q", name)
	}

	task.LookupPath = filepath.Dir(c.File)

	if task.Running {
		return fmt.Errorf("task %s is running, refuse to run another", task.Name)
	}
	err := task.Run(args)
	if err != nil {
		Errorf("error: %s", err)
	}
	task.Running = false
	return nil
}

// Errorf writes to stderr
func Errorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "\n  %s\n\n", fmt.Sprintf(msg, args...))
}

// Fatalf writes to stderr then exit
func Fatalf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\n  %s\n\n", fmt.Sprintf(msg, args...))
	os.Exit(1)
}

// Template helper.
func t(s string) *template.Template {
	return template.Must(template.New("").Funcs(helpers).Parse(s))
}
