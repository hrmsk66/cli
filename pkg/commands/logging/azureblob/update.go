package azureblob

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// UpdateCommand calls the Fastly API to update an Azure Blob Storage logging endpoint.
type UpdateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	NewName           cmd.OptionalString
	AccountName       cmd.OptionalString
	Container         cmd.OptionalString
	SASToken          cmd.OptionalString
	Path              cmd.OptionalString
	Period            cmd.OptionalInt
	GzipLevel         cmd.OptionalInt
	MessageType       cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	ResponseCondition cmd.OptionalString
	TimestampFormat   cmd.OptionalString
	Placement         cmd.OptionalString
	PublicKey         cmd.OptionalString
	FileMaxBytes      cmd.OptionalInt
	CompressionCodec  cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		Manifest: m,
	}
	c.CmdClause = parent.Command("update", "Update an Azure Blob Storage logging endpoint on a Fastly service version")

	// required
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// optional
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("name", "The name of the Azure Blob Storage logging object").Short('n').Required().StringVar(&c.EndpointName)

	// optional
	c.CmdClause.Flag("account-name", "The unique Azure Blob Storage namespace in which your data objects are stored").Action(c.AccountName.Set).StringVar(&c.AccountName.Value)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	c.CmdClause.Flag("container", "The name of the Azure Blob Storage container in which to store logs").Action(c.Container.Set).StringVar(&c.Container.Value)
	c.CmdClause.Flag("file-max-bytes", "The maximum size of a log file in bytes").Action(c.FileMaxBytes.Set).IntVar(&c.FileMaxBytes.Value)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.MessageType(c.CmdClause, &c.MessageType)
	c.CmdClause.Flag("new-name", "New name of the Azure Blob Storage logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	common.Path(c.CmdClause, &c.Path)
	common.Period(c.CmdClause, &c.Period)
	common.Placement(c.CmdClause, &c.Placement)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("sas-token", "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work").Action(c.SASToken.Set).StringVar(&c.SASToken.Value)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.ServiceName.Value,
	})
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateBlobStorageInput, error) {
	input := fastly.UpdateBlobStorageInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}

	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}

	if c.AccountName.WasSet {
		input.AccountName = &c.AccountName.Value
	}

	if c.Container.WasSet {
		input.Container = &c.Container.Value
	}

	if c.SASToken.WasSet {
		input.SASToken = &c.SASToken.Value
	}

	if c.Period.WasSet {
		input.Period = &c.Period.Value
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = &c.GzipLevel.Value
	}

	if c.Format.WasSet {
		input.Format = &c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}

	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = &c.TimestampFormat.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}

	if c.PublicKey.WasSet {
		input.PublicKey = &c.PublicKey.Value
	}

	if c.FileMaxBytes.WasSet {
		input.FileMaxBytes = &c.FileMaxBytes.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	azureblob, err := c.Globals.APIClient.UpdateBlobStorage(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Updated Azure Blob Storage logging endpoint %s (service %s version %d)", azureblob.Name, azureblob.ServiceID, azureblob.ServiceVersion)
	return nil
}
