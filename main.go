package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/schollz/peerdiscovery"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const peerLimit = 1
const peerDiscoveryPort = "30561"
const serverPort = 30562

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
		discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{Limit: peerLimit, Port: peerDiscoveryPort})
		mutex.Lock()
		peers = make([]string, 0)
		for _, d := range discoveries {
			peers = append(peers, d.Address)
			log.Infof("discovered '%s'", d.Address)
		}
		mutex.Unlock()
		log.Info("finished peer discovery")
		time.Sleep(5 * time.Second)
	}
}

func receive(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	text := fmt.Sprintf("%s", data)
	if text != "" {
		err := clipboard.WriteAll(text)
		if err != nil {
			log.WithError(err).Fatal("can't write into clipboard")
		}
	}
	fmt.Fprintf(w, "received: '%s'\n", text)
}

func startServer() {
	http.HandleFunc("/receive", receive)
	http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)
}

func main() {
	clipboardContent := make(chan string)
	go captureClipboard(clipboardContent)
	go startDiscovery()
	go startServer()

	for value := range clipboardContent {
		for _, peer := range peers {
			mutex.Lock()
			log.Infof("sending %s to peer %s", value, peer)
			mutex.Unlock()
		}
	}
}
