package application

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/codegangsta/cli"
	"time"
)

type Logs struct {
	ui       terminal.UI
	config   *configuration.Configuration
	logsRepo api.LogsRepository
	appReq   requirements.ApplicationRequirement
}

func NewLogs(ui terminal.UI, config *configuration.Configuration, logsRepo api.LogsRepository) (cmd *Logs) {
	cmd = new(Logs)
	cmd.ui = ui
	cmd.config = config
	cmd.logsRepo = logsRepo
	return
}

func (cmd *Logs) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c, "logs")
		err = errors.New("Incorrect Usage")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *Logs) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	logChan := make(chan *logmessage.Message, 1000)
	var err error

	if c.Bool("recent") {
		onConnect := func() {
			cmd.ui.Say("Connected, dumping recent logs for app %s in org %s / space %s as %s...\n",
				terminal.EntityNameColor(app.Name),
				terminal.EntityNameColor(cmd.config.Organization.Name),
				terminal.EntityNameColor(cmd.config.Space.Name),
				terminal.EntityNameColor(cmd.config.Username()),
			)
		}
		err = cmd.logsRepo.RecentLogsFor(app, onConnect, logChan)
	} else {

		onConnect := func() {
			cmd.ui.Say("Connected, tailing logs for app %s in org %s / space %s as %s...\n",
				terminal.EntityNameColor(app.Name),
				terminal.EntityNameColor(cmd.config.Organization.Name),
				terminal.EntityNameColor(cmd.config.Space.Name),
				terminal.EntityNameColor(cmd.config.Username()),
			)
		}
		// in this case we tail the logs forever, so we never send true on this channel
		stopLoggingChan := make(chan bool)

		err = cmd.logsRepo.TailLogsFor(app, onConnect, logChan, stopLoggingChan, 5*time.Second)
	}
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.displayLogMessages(logChan)
}

func (cmd *Logs) displayLogMessages(logChan chan *logmessage.Message) {
	for msg := range logChan {
		cmd.ui.Say(logMessageOutput(msg))
	}
}
