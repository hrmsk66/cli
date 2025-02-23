package dictionaryentry

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// UpdateCommand calls the Fastly API to update a dictionary item.
type UpdateCommand struct {
	cmd.Base

	Input       fastly.UpdateDictionaryItemInput
	InputBatch  fastly.BatchModifyDictionaryItemsInput
	file        cmd.OptionalString
	manifest    manifest.Data
	serviceName cmd.OptionalServiceNameID
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("update", "Update or insert an item on a Fastly edge dictionary")

	// required
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)

	// optional
	c.CmdClause.Flag("file", "Batch update json file").Action(c.file.Set).StringVar(&c.file.Value)
	c.CmdClause.Flag("key", "Dictionary item key").StringVar(&c.Input.ItemKey)
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
	c.CmdClause.Flag("value", "Dictionary item value").StringVar(&c.Input.ItemValue)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID
	c.InputBatch.ServiceID = serviceID
	c.InputBatch.DictionaryID = c.Input.DictionaryID

	if c.file.WasSet {
		err := c.batchModify(out)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		return nil
	}

	if c.Input.ItemKey == "" || c.Input.ItemValue == "" {
		return fmt.Errorf("an empty value is not allowed for either the '--key' or '--value' flags")
	}

	d, err := c.Globals.APIClient.UpdateDictionaryItem(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated dictionary item (service %s)", d.ServiceID)
	text.Break(out)
	text.PrintDictionaryItem(out, "", d)
	return nil
}

func (c *UpdateCommand) batchModify(out io.Writer) error {
	jsonFile, err := os.Open(c.file.Value)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	err = json.Unmarshal(jsonBytes, &c.InputBatch)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if len(c.InputBatch.Items) == 0 {
		return fmt.Errorf("item key not found in file %s", c.file.Value)
	}

	err = c.Globals.APIClient.BatchModifyDictionaryItems(&c.InputBatch)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Made %d modifications of Dictionary %s on service %s", len(c.InputBatch.Items), c.Input.DictionaryID, c.InputBatch.ServiceID)
	return nil
}
