package main

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"filippo.io/age"
	"github.com/schollz/peerdiscovery"
	log "github.com/sirupsen/logrus"
	"golang.design/x/clipboard"
)

const peerDiscoveryPort = "30561"
const serverPort = 30562
const username = "clipu"

var password = ""
var peerLimit = 1

var peers = make([]string, 0)
var mutex = &sync.Mutex{}
var lastReceived = ""
var allowSelf = false

func captureClipboard(clipboardContents chan<- string) {
	previousContent := ""
	for {
		currentContent := string(clipboard.Read(clipboard.FmtText))
		if previousContent != currentContent && lastReceived != currentContent {
			log.Debugf("got text '%s'", currentContent)
			previousContent = currentContent
			clipboardContents <- currentContent
		}
		time.Sleep(1 * time.Second)
	}
}

func startDiscovery() {
	for {
		log.Debug("started peer discovery")
		discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{Limit: peerLimit, Port: peerDiscoveryPort, AllowSelf: allowSelf})

		newPeers := make([]string, 0)
		for _, d := range discoveries {
			if authorized(d.Address) {
				newPeers = append(newPeers, d.Address)
				log.Debugf("discovered '%s'", d.Address)
			}
		}

		mutex.Lock()
		peers = newPeers
		mutex.Unlock()
		log.Debug("finished peer discovery")
		time.Sleep(5 * time.Second)
	}
}

func basicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		handler(w, r)
	}
}

func encryptText(text string) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(w, text); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decryptText(data []byte) (string, error) {
	identity, err := age.NewScryptIdentity(password)
	if err != nil {
		return "", err
	}
	r, err := age.Decrypt(bytes.NewReader(data), identity)
	if err != nil {
		return "", err
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func receive(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := io.ReadAll(r.Body)
	if len(data) == 0 {
		return
	}
	text, err := decryptText(data)
	if err != nil {
		log.WithError(err).Error("failed to decrypt received data")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if text != "" {
		log.Infof("received %s", text)
		clipboard.Write(clipboard.FmtText, []byte(text))
		lastReceived = text
	}
	fmt.Fprintf(w, "received: '%s'\n", text)
}

func ack(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Fprintf(w, "ok\n")
}

func authorized(peer string) bool {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/ack", peer, serverPort), nil)
	if err != nil {
		log.WithError(err).Errorf("could not create request to ask for authorization")
		return false
	}
	req.SetBasicAuth(username, password)
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	if res.StatusCode != http.StatusOK {
		return false
	}
	log.Infof("peer %s responded positively to authorization inquiry", peer)
	return true
}

func startServer() {
	http.HandleFunc("/receive", basicAuth(receive, username, password, ""))
	http.HandleFunc("/ack", basicAuth(ack, username, password, ""))
	http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)
}

func sendText(text string, peer string) error {
	encrypted, err := encryptText(text)
	if err != nil {
		return fmt.Errorf("failed to encrypt text: %w", err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/receive", peer, serverPort), bytes.NewReader(encrypted))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("peer has returned an invalid code %d", res.StatusCode)
	}
	return nil
}

func init() {
	if levelStr := os.Getenv("CLIPU_LOG_LEVEL"); levelStr != "" {
		level, err := log.ParseLevel(levelStr)
		if err != nil {
			log.WithError(err).Fatalf("invalid log level %q", levelStr)
		}
		log.SetLevel(level)
	}

	err := clipboard.Init()
	if err != nil {
		log.WithError(err).Fatal("can't initialize clipboard")
	}
	text := string(clipboard.Read(clipboard.FmtText))
	lastReceived = text

	password = os.Getenv("CLIPU_PASSWORD")
	if password == "" {
		log.Fatal("running on empty password! Set CLIPU_PASSWORD to start.")
	}
	if _, found := os.LookupEnv("CLIPU_ALLOW_SELF"); found {
		allowSelf = true
	}

	peerLimitStr, found := os.LookupEnv("CLIPU_PEER_LIMIT")
	if found {
		peerLimitInt, err := strconv.Atoi(peerLimitStr)
		if err != nil {
			log.WithError(err).Errorf("Can't parse peer limit from %s", peerLimitStr)
		} else {
			peerLimit = peerLimitInt
		}
	}
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
			err := sendText(value, peer)
			if err != nil {
				log.WithError(err).Errorf("an error has occurred while sending text %s to %s", value, peer)
			}
			mutex.Unlock()
		}
	}
}
