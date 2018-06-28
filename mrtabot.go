package mmrtabot

import (
	"time"

	"github.com/asdine/storm"
	"github.com/konkers/mmrta"
	"github.com/konkers/teletran"
)

type MmrtabotModule struct {
	bot    *teletran.Bot
	db     storm.Node
	client *mmrta.Client

	ticker *time.Ticker
}

func NewMmrtabotModule(bot *teletran.Bot) (*MmrtabotModule, error) {

	c, err := mmrta.NewClient()
	if err != nil {
		return nil, err
	}

	module := &MmrtabotModule{
		bot:    bot,
		db:     bot.GetDbBucket("mmrta"),
		client: c,
	}

	bot.AddCommand("backlog", "Report verification backlog.", module.backlogCommand)

	bot.AddAdminCommand("announcements", "Control announcements.", module.announcementsCommand)

	module.ticker = time.NewTicker(1 * time.Minute)
	go module.tickerHandler()

	return module, nil
}
