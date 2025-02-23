package s3

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

// ListCommand calls the Fastly API to list Amazon S3 logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListS3sInput
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
	c.CmdClause = parent.Command("list", "List S3 endpoints on a Fastly service version")

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

	s3s, err := c.Globals.APIClient.ListS3s(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(s3s)
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
		for _, s3 := range s3s {
			tw.AddLine(s3.ServiceID, s3.ServiceVersion, s3.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, s3 := range s3s {
		fmt.Fprintf(out, "\tS3 %d/%d\n", i+1, len(s3s))
		fmt.Fprintf(out, "\t\tService ID: %s\n", s3.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", s3.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", s3.Name)
		fmt.Fprintf(out, "\t\tBucket: %s\n", s3.BucketName)
		if s3.AccessKey != "" || s3.SecretKey != "" {
			fmt.Fprintf(out, "\t\tAccess key: %s\n", s3.AccessKey)
			fmt.Fprintf(out, "\t\tSecret key: %s\n", s3.SecretKey)
		}
		if s3.IAMRole != "" {
			fmt.Fprintf(out, "\t\tIAM role: %s\n", s3.IAMRole)
		}
		fmt.Fprintf(out, "\t\tPath: %s\n", s3.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", s3.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", s3.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", s3.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", s3.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", s3.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", s3.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", s3.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", s3.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", s3.PublicKey)
		fmt.Fprintf(out, "\t\tRedundancy: %s\n", s3.Redundancy)
		fmt.Fprintf(out, "\t\tServer-side encryption: %s\n", s3.ServerSideEncryption)
		fmt.Fprintf(out, "\t\tServer-side encryption KMS key ID: %s\n", s3.ServerSideEncryption)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", s3.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
