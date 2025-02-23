package config

import (
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	location bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("config", "Display the Fastly CLI configuration")
	c.CmdClause.Flag("location", "Print the location of the CLI configuration file").Short('l').BoolVar(&c.location)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	if c.location {
		if c.Globals.Flags.Verbose {
			text.Break(out)
		}
		fmt.Fprintln(out, c.Globals.Path)
		return nil
	}

	data, err := os.ReadFile(c.Globals.Path)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	fmt.Fprintln(out, string(data))
	return nil
}
