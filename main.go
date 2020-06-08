package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

// TODO:
// bot struct + constructor
// add timeout ctx
// docker image + keybase login
// loop/bot detection

func fail(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(3)
}

func logerr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func main() {
	const msgFilter = "https://twitter.com/"

	var kbLoc string
	var kbc *kbchat.API
	var err error

	flag.StringVar(&kbLoc, "keybase", "keybase", "the location of the Keybase app")
	flag.Parse()

	if kbc, err = kbchat.Start(
		kbchat.RunOptions{KeybaseLocation: kbLoc},
	); err != nil {
		fail("Error creating API: %s", err.Error())
	}

	commands := []chat1.UserBotCommandInput{
		{
			Name:        "tw",
			Description: "get the preview of a twitter link",
			Usage:       "!tw [link]",
			ExtendedDescription: &chat1.UserBotExtendedDescription{
				Title:       "tw command",
				DesktopBody: "tw",
				MobileBody:  "tw",
			},
		},
	}
	if _, err := kbc.AdvertiseCommands(kbchat.Advertisement{
		Advertisements: []chat1.AdvertiseCommandAPIParam{
			{
				Typ:      "public",
				Commands: commands,
			},
		},
	}); err != nil {
		fail("Error creating bot: %s", err.Error())
	}

	sub, err := kbc.ListenForNewTextMessages()
	if err != nil {
		fail("Error listening: %s", err.Error())
	}

	for {
		msg, err := sub.Read()
		if err != nil {
			logerr("failed to read message: %s", err.Error())
		}

		if msg.Message.Content.TypeName != "text" {
			continue
		}

		// TODO: Botinfo cannot be used :/
		if msg.Message.Sender.Username == "tweeto" {
			logerr("message comes from myself: %s", msg.Message.Sender)
			continue
		}

		if !strings.Contains(msg.Message.Content.Text.Body, msgFilter) {
			logerr("Message doesn't pass filter: %s", msg.Message.Content.Text.Body)
			continue
		}

		url, err := getURLFromBody(msg.Message.Content.Text.Body)
		if err != nil {
			logerr("no URL found in message: %s", err.Error())
			continue
		}

		resp, err := getPreviewFromURL(url)
		if err != nil {
			logerr("cannot get preview: %s", err.Error())
			continue
		}

		if _, err = kbc.SendMessageByConvID(msg.Message.ConvID, resp); err != nil {
			fail("error sending response: %s", err.Error())
		}

	}

}
