package cli

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Description describes an entity in the CLI
type Description struct {
	// Name of this thing
	Name string
	// Short, one-sentence description
	Short string
	// Longer, multi-line description
	Long string
	// Optional grouping information
	Tags []string
	// Original name, before mangling.
	Original string
}

// Doc outputs the documentation for this description
func (desc *Description) Doc(out io.Writer) {
	fmt.Fprintf(out, "%s - %s\n", desc.Name, desc.Short)
	if desc.Long != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, desc.Long)
	}
	if len(desc.Tags) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Tags: ")
		fmt.Fprintln(out, "  "+strings.Join(desc.Tags, ", "))
	}
}

// Command represents a single callable command
type Command struct {
	Description
	// Optional environment variables
	Optional []string
	// Required environment variables
	Required []string
}

// Arg represents a single environment variable with special meaning
type Arg struct {
	Description
}

// UI is a user interface
type UI struct {
	Description
	Commands []Command
	Args     []Arg
}

// Run executes the commands specified by argv[] by calling the methods with
// those names on `impl`.
//
// RunCommands will panic if any error is encountered.
func (ui *UI) Run(
	// This function will be called by Run() to find the implemenation for a command name.
	// You can use the provided ui.DynamicCommandLookup(impl), or you can generate a
	// completely type-safe UI, and use ui.staticCommandLookup.
	getCommandMethod func(commandName string) (impl func(), found bool),
	commandNames []string,
) {
	unknown := make([]string, 0)
	queue := make([]func(), 0, len(commandNames))
	for _, n := range commandNames {
		fn, found := getCommandMethod(n)
		if !found {
			// special case for "help": provide help on any other commands given and
			// do nothing else
			if n == "help" {
				queue = []func(){ui.HelpFor(commandNames)}
				break
			}
			unknown = append(unknown, n)
			continue
		}
		queue = append(queue, fn)
	}

	if len(unknown) > 0 {
		panic(fmt.Sprintf("Unknown commands: %v", unknown))
	}

	for _, fn := range queue {
		fn()
	}
}

// DynamicCommandLookup returns a function for Run() that looks up the
// implementation for a command based on the command's Original field.
//
// This operation is unsafe, and could panic due to type coercion at runtime.
// This package provides a tool for generating a type-safe alternative.
func (ui *UI) DynamicCommandLookup(impl interface{}) func(string) (func(), bool) {
	t := reflect.TypeOf(impl)
	return func(name string) (func(), bool) {
		cmd := ui.GetCommand(name)
		if cmd == nil {
			return nil, false
		}
		fn, found := t.MethodByName(cmd.Original)
		if !found {
			return nil, found
		}
		return fn.Func.Interface().(func()), true
	}
}

func (ui *UI) HelpFor(commandNames []string) func() {
	return func() {
		if len(commandNames) == 0 || (len(commandNames) == 1 && commandNames[0] == "help") {
			ui.Overview(os.Stdout)
			return
		}

		for _, name := range commandNames {
			// skip help - handled above
			if name == "help" {
				continue
			}
			err := ui.AboutCommand(name, os.Stdout)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

func (ui *UI) namePadding() int {
	maxlen := 0
	for _, d := range ui.Commands {
		if maxlen < len(d.Name) {
			maxlen = len(d.Name)
		}
	}
	for _, d := range ui.Args {
		if maxlen < len(d.Name) {
			maxlen = len(d.Name)
		}
	}
	return maxlen
}

func (ui *UI) shortFormat() string {
	l := ui.namePadding()
	return fmt.Sprintf("  %%%ds    %%s\n", l)
}

func (ui *UI) Overview(out io.Writer) {
	fmt.Fprintf(out, "%s - %s\n", ui.Name, ui.Short)
	fmt.Fprintln(out, "")
	if ui.Long != "" {
		fmt.Fprintln(out, ui.Long)
		fmt.Fprintln(out, "")
	}
	fmt.Fprintln(out, "Commands:")
	freq := make(map[string]int)
	format := ui.shortFormat()
	for _, cmd := range ui.Commands {
		fmt.Fprintf(out, format, cmd.Name, cmd.Short)
		for _, name := range cmd.Optional {
			freq[name] = freq[name] + 1
		}
		for _, name := range cmd.Required {
			freq[name] = freq[name] + 1
		}
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Common Arguments:")
	for _, arg := range ui.Args {
		if freq[arg.Name] == 0 {
			continue
		}
		fmt.Fprintf(out, format, arg.Name, arg.Short)
	}
}

func (ui *UI) GetArg(name string) *Arg {
	for _, arg := range ui.Args {
		if arg.Name == name {
			return &arg
		}
	}
	return nil
}

func (ui *UI) GetCommand(name string) *Command {
	for _, cmd := range ui.Commands {
		if cmd.Name == name {
			return &cmd
		}
	}
	return nil
}

func (ui *UI) AboutCommand(name string, out io.Writer) error {
	cmd := ui.GetCommand(name)
	if cmd == nil {
		return fmt.Errorf("Unknown command %q", name)
	}

	cmd.Doc(out)

	if len(cmd.Required) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Required Arguments:")
		format := ui.shortFormat()
		for _, name := range cmd.Required {
			arg := ui.GetArg(name)
			fmt.Fprintf(out, format, arg.Name, arg.Short)
		}
	}

	if len(cmd.Optional) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Optional Arguments:")
		format := ui.shortFormat()
		for _, name := range cmd.Optional {
			arg := ui.GetArg(name)
			fmt.Fprintf(out, format, arg.Name, arg.Short)
		}
	}

	return nil
}

func (ui *UI) AboutArg(name string, out io.Writer) error {
	arg := ui.GetArg(name)
	if arg == nil {
		return fmt.Errorf("Unknown argument %q", name)
	}
	arg.Doc(out)
	return nil
}
