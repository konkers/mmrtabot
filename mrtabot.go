package mmrtabot

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/konkers/mmrta"
	"github.com/konkers/teletran"
	"github.com/olekukonko/tablewriter"
)

type MmrtabotModule struct {
	bot    *teletran.Bot
	client *mmrta.Client
}

func NewMmrtabotModule(bot *teletran.Bot) (*MmrtabotModule, error) {

	c, err := mmrta.NewClient()
	if err != nil {
		return nil, err
	}

	module := &MmrtabotModule{
		bot:    bot,
		client: c,
	}

	bot.AddCommand("backlog", "Report verification backlock.", module.backlogCommand)

	return module, nil
}

func (m *MmrtabotModule) backlogCommand(ctx *teletran.CommandContext, args []string) {
	runs, err := m.client.GetUnverifiedRuns(true)
	if err != nil {
		ctx.SendResponse(fmt.Sprintf("Can't get runs: %s", err.Error()))
		return
	}

	sort.Slice(runs, func(i, j int) bool { return runs[i].Id < runs[j].Id })

	writer := bytes.NewBufferString("```")
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Time", "Game", "Category", "Runner"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	for _, r := range runs {
		table.Append([]string{
			r.Duration().String(),
			r.Game.AbbrevName(),
			r.AbbrevCat(),
			r.User.Name,
		})
	}
	table.Render()
	ctx.SendResponse(writer.String() + "```")

}
