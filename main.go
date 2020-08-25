package main

import (
	"github.com/atotto/clipboard"
	"github.com/schollz/peerdiscovery"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const peerLimit = 1

var peers = make([]string, 0)
var mutex = &sync.Mutex{}

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

func startDiscovery() {
	for {
		log.Info("started peer discovery")
		discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{Limit: peerLimit})
		mutex.Lock()
		peers = make([]string, 0)
		for _, d := range discoveries {
			peers = append(peers, d.Address)
			log.Infof("discovered '%s'\n", d.Address)
		}
		mutex.Unlock()
		log.Info("started peer discovery")
		time.Sleep(5 * time.Second)
	}
}

func main() {
	clipboardContent := make(chan string)
	go captureClipboard(clipboardContent)
	go startDiscovery()

	for value := range clipboardContent {
		for _, peer := range peers {
			mutex.Lock()
			log.Infof("sending %s to peer %s", value, peer)
			mutex.Unlock()
		}
	}
}
