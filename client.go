package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sabhiram/go-wol/wol"
)

const (
	resDone  = "done"
	resError = "error"
)

var (
	hostAddr string
	secure   bool
)

func client(pass string) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	scheme := func() string {
		if secure {
			return "wss"
		}
		return "ws"
	}()
	u := url.URL{
		Scheme: scheme,
		Host:   hostAddr,
		Path:   fmt.Sprintf("/%s/streaming", pass),
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	request := make(chan []byte)
	go func() {
		defer close(request)
		log.Println("Ready...")
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Fatal(err)
				return
			}
			request <- msg
		}
	}()

	for {
		select {
		case req := <-request:
			switch {
			case string(req) == cmdInfo:
				log.Println("Information requested")
				if err := c.WriteMessage(websocket.TextMessage, []byte("available")); err != nil {
					return err
				}
			case strings.HasPrefix(string(req), cmdWake):
				if err := func() error {
					addr := strings.Split(string(req), " ")[1]
					log.Printf("Waking %s\n", addr)
					mp, err := wol.New(addr)
					if err != nil {
						return err
					}

					bytes, err := mp.Marshal()
					if err != nil {
						return err
					}

					bcAddr := fmt.Sprintf("%s:%s", "255.255.255.255", "9")
					uAddr, err := net.ResolveUDPAddr("udp", bcAddr)
					if err != nil {
						return err
					}

					udpConn, err := net.DialUDP("udp", nil, uAddr)
					if err != nil {
						return err
					}
					defer udpConn.Close()

					n, err := udpConn.Write(bytes)
					if err == nil && n != 102 {
						err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
					}
					if err != nil {
						return err
					}

					return nil
				}(); err != nil {
					if err := c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s %s", resError, err.Error()))); err != nil {
						return err
					}
				} else {
					if err := c.WriteMessage(websocket.TextMessage, []byte(resDone)); err != nil {
						return err
					}
				}
			}
		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "interrupt"))
			if err != nil {
				return err
			}
			select {
			case <-request:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}
