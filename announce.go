package mmrtabot

import (
	"fmt"
	"log"
	"time"

	"github.com/asdine/storm"
	"github.com/konkers/teletran"
	"github.com/olekukonko/tablewriter"
)

type announceConfig struct {
	ChannelID   string `storm:"id"`
	Enabled     bool
	NextMessage time.Time
	Period      time.Duration
}

func (m *MmrtabotModule) tickerHandler() {
	for {
		select {
		case t := <-m.ticker.C:
			m.handleAnnounce(t)
		}
	}
}

func (m *MmrtabotModule) handleAnnounce(t time.Time) {

	configs, err := m.getAnnounceConfigs()
	if err != nil {
		log.Printf("can't get announce configs: %v\n", err)
		return
	}

	var fetchedBacklog = false
	var msg = ""
	for _, config := range configs {
		if config.Enabled && t.After(config.NextMessage) {
			if !fetchedBacklog {
				var err error
				msg, err = m.backlogMessage()
				if err != nil {
					log.Printf("can't get backlog message: %v\n", err)
					return
				}
				fetchedBacklog = true
			}
			if msg != "" {
				m.bot.Session.ChannelMessageSend(config.ChannelID, msg)
			} else {
				c, err := m.bot.Session.Channel(config.ChannelID)
				if err == nil {
					log.Printf("No runs to announce on %s", c.Name)
				}
			}
			config.NextMessage = config.NextMessage.Add(config.Period)
			m.announceBucket().Save(&config)
		}
	}
}

func (m *MmrtabotModule) announceBucket() storm.Node {
	return m.db.From("announcements")
}

func (m *MmrtabotModule) getAnnounceConfigs() ([]announceConfig, error) {
	var configs []announceConfig
	err := m.announceBucket().All(&configs)

	return configs, err
}

func (m *MmrtabotModule) announcementsCommandList(ctx *teletran.CommandContext) {
	configs, err := m.getAnnounceConfigs()
	if err != nil {
		fmt.Fprintf(ctx, "can't get announce configs: %v\n", err)
		return
	}

	table := tablewriter.NewWriter(ctx)
	table.SetHeader([]string{"Server", "Channel", "Period", "Next Message"})
	for _, config := range configs {
		c, err := ctx.Session.Channel(config.ChannelID)
		if err != nil {
			fmt.Fprintf(ctx, "can't get channel: %v\n", config.ChannelID)
			continue
		}

		s, err := ctx.Session.Guild(c.GuildID)
		if err != nil {
			fmt.Fprintf(ctx, "can't get server for: %v\n", c.Name)
			continue
		}

		if config.Enabled {
			table.Append([]string{
				s.Name,
				c.Name,
				config.Period.String(),
				config.NextMessage.Format(time.RFC1123),
			})
		}
	}
	fmt.Fprintf(ctx, "```")
	table.Render()
	fmt.Fprintf(ctx, "```")
}

func (m *MmrtabotModule) announcementsCommandEnable(ctx *teletran.CommandContext, enable bool, period time.Duration) {
	config := announceConfig{
		ChannelID:   ctx.Message.ChannelID,
		Enabled:     enable,
		Period:      period,
		NextMessage: time.Now(),
	}
	err := m.announceBucket().Save(&config)
	if err != nil {
		fmt.Fprintf(ctx, "error: %v\n", err)
	}

	if enable {
		fmt.Fprintf(ctx, "I'll announce the verification backog every %v, starting now.\n", period)
	} else {
		fmt.Fprintf(ctx, "Announcements disabled for this channel.\n")
	}
}

func (m *MmrtabotModule) announcementsCommand(ctx *teletran.CommandContext, args []string) {
	if len(args) == 0 {
		m.announcementsCommandList(ctx)
		return
	}

	switch args[0] {
	case "on":
		d := 12 * time.Hour
		if len(args) == 2 {
			var err error
			d, err = time.ParseDuration(args[1])
			if err != nil {
				fmt.Fprintf(ctx, "Can not parse duration of %s.", args[1])
				return
			}
		}
		m.announcementsCommandEnable(ctx, true, d)
	case "off":
		m.announcementsCommandEnable(ctx, false, 0)
	default:
		fmt.Fprintf(ctx, "```\n")
		fmt.Fprintf(ctx, "usage: announcements <cmd>\n")
		fmt.Fprintf(ctx, "  commands:\n")
		fmt.Fprintf(ctx, "    on [period] - enable announcements for this channel.\n")
		fmt.Fprintf(ctx, "    off         - disable announcements for this channel.\n")
		fmt.Fprintf(ctx, "```\n")
	}
}
