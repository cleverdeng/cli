package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Stacks struct {
	ui         terminal.UI
	config     *configuration.Configuration
	stacksRepo api.StackRepository
}

func NewStacks(ui terminal.UI, config *configuration.Configuration, stacksRepo api.StackRepository) (cmd *Stacks) {
	cmd = new(Stacks)
	cmd.ui = ui
	cmd.config = config
	cmd.stacksRepo = stacksRepo
	return
}

func (cmd *Stacks) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *Stacks) Run(c *cli.Context) {
	cmd.ui.Say("Getting stacks in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	stacks, apiResponse := cmd.stacksRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		[]string{"name", "description"},
	}

	for _, stack := range stacks {
		table = append(table, []string{
			stack.Name,
			stack.Description,
		})
	}

	cmd.ui.DisplayTable(table)
}
