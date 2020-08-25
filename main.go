package main

import (
	"github.com/atotto/clipboard"
	"github.com/schollz/peerdiscovery"
	log "github.com/sirupsen/logrus"
	"time"
)

func captureClipboard(clipboardContents chan<- string) {
	previousContent := ""
	for {
		currentContent, err := clipboard.ReadAll()
		if err != nil {
			log.WithError(err).Fatal("can't read clipboard content")
		}
		if previousContent != currentContent {
			log.Infof("got text '%s'", currentContent)
			previousContent = currentContent
			clipboardContents <- currentContent
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	clipboardContent := make(chan string)
	go captureClipboard(clipboardContent)

	for value := range clipboardContent {
		discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{Limit: 1})
		for _, d := range discoveries {
			log.Infof("discovered '%s'\n", d.Address)
			log.Infof("sending %s", value)
		}
	}
}
