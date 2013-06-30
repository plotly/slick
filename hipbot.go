package main

import (
	"flag"
	"github.com/tkawachi/hipbot/inout"
	"github.com/tkawachi/hipbot/plugin"
	"github.com/tkawachi/hipchat"
	"log"
	"os"
	"strings"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.hipbot", "config file")

const (
	ConfDomain = "conf.hipchat.com"
)

func registerPlugins() []plugin.Plugin {
	plugins := make([]plugin.Plugin, 0)
	plugins = append(plugins, new(Wikipedia))
	plugins = append(plugins, inout.New())
	return plugins
}

func replyHandler(client *hipchat.Client, nickname string, respCh chan *plugin.HandleReply) {
	for {
		reply := <-respCh
		if reply != nil {
			client.Say(reply.To, nickname, reply.Message)
		}
	}
}

func messageHandler(client *hipchat.Client, plugins []plugin.Plugin, respCh chan *plugin.HandleReply) {
	msgs := client.Messages()
	for {
		msg := <-msgs
		for _, p := range plugins {
			go func(p plugin.Plugin) { respCh <- p.Handle(msg) }(p)
		}
	}
}

func main() {
	flag.Parse()
	config := loadConfig(*configFile)
	plugins := registerPlugins()
	chatConfig := config.Hipchat
	client, err := hipchat.NewClient(
		chatConfig.Username, chatConfig.Password, chatConfig.Resource)
	if err != nil {
		log.Fatal(err)
	}
	for _, room := range chatConfig.Rooms {
		if !strings.Contains(room, "@") {
			room = room + "@" + ConfDomain
		}
		client.Join(room, chatConfig.Nickname)
	}
	respCh := make(chan *plugin.HandleReply)
	go client.KeepAlive()
	go replyHandler(client, chatConfig.Nickname, respCh)
	go messageHandler(client, plugins, respCh)
	log.Println("hipbot started")
	select {}
}
