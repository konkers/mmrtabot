package mmrtabot

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/konkers/teletran"
	"github.com/olekukonko/tablewriter"
)

func (m *MmrtabotModule) backlogMessage() (string, error) {
	runs, err := m.client.GetUnverifiedRuns(true)
	if err != nil {
		return "", err
	}

	if len(runs) == 0 {
		return "", nil
	}

	sort.Slice(runs, func(i, j int) bool { return runs[i].Id < runs[j].Id })

	header := fmt.Sprintf("```run backlog as of %s:\n", time.Now().Format(time.RFC1123))
	writer := bytes.NewBufferString(header)
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Time", "Game", "Category", "Runner"})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT})
	table.SetColWidth(20)
	for _, r := range runs {
		table.Append([]string{
			r.PrettyTime(),
			r.Game.AbbrevName(),
			r.AbbrevCat(),
			r.User.Name,
		})
	}
	table.Render()

	return writer.String() + "```", nil
}

func (m *MmrtabotModule) backlogCommand(ctx *teletran.CommandContext, args []string) {
	msg, err := m.backlogMessage()
	if err != nil {
		fmt.Fprintf(ctx, "Can't get runs: %v", err)
		return
	}
	if msg == "" {
		fmt.Fprintf(ctx, "No backlogged runs.\n")
	} else {
		ctx.SendResponse(msg)
	}

}
