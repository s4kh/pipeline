package trades

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/s4kh/trader-app/producer/msgbroker"
)

const (
	subscribeId = iota
	unsubscribeId
)

const host = "stream.binance.com:443"

var conn *websocket.Conn

type message struct {
	Id     int      `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

func getConnection() (*websocket.Conn, error) {
	if conn != nil {
		return conn, nil
	}

	u := url.URL{Scheme: "wss", Host: host, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	dialer := websocket.DefaultDialer

	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error during ws dial: statusCode: %v, %v", resp.StatusCode, err)
	}

	conn = c

	return conn, nil
}

func unsubscribeOnClose(conn *websocket.Conn, tradeTopics []string) error {
	msg := &message{Id: unsubscribeId, Method: "UNSUBSCRIBE", Params: tradeTopics}

	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error during msg marshal in unsubscribeonclose: %v", err)
	}

	conn.WriteMessage(websocket.TextMessage, b)

	return nil
}

func pongHandler(conn *websocket.Conn) func(s string) error {
	return func(appData string) error {
		log.Println("Received pong:", appData)

		pingFrame := []byte{1, 2, 3, 4, 5}

		err := conn.WriteMessage(websocket.PingMessage, pingFrame)
		if err != nil {
			return fmt.Errorf("failed to send ping msg: %v", err)
		}

		return nil
	}
}

func CloseConnections() {
	conn.Close()
}

func SubscribeAndListen(topics []string, mb msgbroker.MsgBroker) error {
	conn, err := getConnection()
	if err != nil {
		return fmt.Errorf("could not get connection: %v", err)
	}

	conn.SetPongHandler(pongHandler(conn))

	tradeTopics := make([]string, 0, len(topics))

	for _, t := range topics {
		tradeTopics = append(tradeTopics, t+"@aggTrade")
	}
	defer unsubscribeOnClose(conn, tradeTopics)

	msg := &message{Id: subscribeId, Method: "SUBSCRIBE", Params: tradeTopics}

	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error during subscribe msg marshal: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		return fmt.Errorf("could not send subscribe msg: %v", err)
	}

	log.Println("listening", tradeTopics)

	resChan := make(chan msgbroker.PublishRes)

	for {
		t, payload, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("could not read ws message: %v", err)
		}

		trade := Ticker{}

		err = json.Unmarshal(payload, &trade)
		if err != nil {
			return fmt.Errorf("could not unmarshal trade msg: %v", err)
		}
		// first readmessage returns
		if trade.Symbol == "" {
			fmt.Println(t)
			continue
		}

		log.Println("GOT ", trade.Symbol, trade.Price, trade.Quantity)

		go listenWorkerRes(resChan)

		go func() {
			publishToMsgBroker(mb, trade, resChan)
		}()
	}
}

func listenWorkerRes(resChan <-chan msgbroker.PublishRes) {
	for val := range resChan {
		fmt.Println(val)
	}
}

func publishToMsgBroker(mb msgbroker.MsgBroker, t Ticker, resChan chan<- msgbroker.PublishRes) {
	key := fmt.Sprintf("%s-%s", t.Symbol, strconv.Itoa(int(t.Time)))
	topic := fmt.Sprintf("trades-%s", t.Symbol)
	mb.Publish(t.String(), key, topic, resChan)
}
