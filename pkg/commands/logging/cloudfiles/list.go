package cloudfiles

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// ListCommand calls the Fastly API to list Cloudfiles logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListCloudfilesInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("list", "List Cloudfiles endpoints on a Fastly service version")

	// required
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// optional
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	cloudfiles, err := c.Globals.APIClient.ListCloudfiles(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(cloudfiles)
			if err != nil {
				return err
			}
			_, err = out.Write(data)
			if err != nil {
				c.Globals.ErrLog.Add(err)
				return fmt.Errorf("error: unable to write data to stdout: %w", err)
			}
			return nil
		}

		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, cloudfile := range cloudfiles {
			tw.AddLine(cloudfile.ServiceID, cloudfile.ServiceVersion, cloudfile.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, cloudfile := range cloudfiles {
		fmt.Fprintf(out, "\tCloudfiles %d/%d\n", i+1, len(cloudfiles))
		fmt.Fprintf(out, "\t\tService ID: %s\n", cloudfile.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", cloudfile.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", cloudfile.Name)
		fmt.Fprintf(out, "\t\tUser: %s\n", cloudfile.User)
		fmt.Fprintf(out, "\t\tAccess key: %s\n", cloudfile.AccessKey)
		fmt.Fprintf(out, "\t\tBucket: %s\n", cloudfile.BucketName)
		fmt.Fprintf(out, "\t\tPath: %s\n", cloudfile.Path)
		fmt.Fprintf(out, "\t\tRegion: %s\n", cloudfile.Region)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", cloudfile.Placement)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", cloudfile.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", cloudfile.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", cloudfile.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", cloudfile.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", cloudfile.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", cloudfile.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", cloudfile.TimestampFormat)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", cloudfile.PublicKey)
	}
	fmt.Fprintln(out)

	return nil
}
