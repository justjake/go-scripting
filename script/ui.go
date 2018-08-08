package script

import (
	"fmt"
	"io"
	"strings"
)

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

func (ui *UI) AboutCommand(name string, out io.Writer) error {
	var cmd *Command
	for _, maybe := range ui.Commands {
		if maybe.Name == name {
			cmd = &maybe
		}
	}
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
