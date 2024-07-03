package gate

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Websocket interface {
	Connect(out bool) error
	Close() error
	SubOrderBook(cp, level, interval string) error
	Read() interface{}
	Message() chan interface{}
}

type wsClient struct {
	url    string
	cli    *websocket.Conn
	stop   chan interface{}
	read   chan interface{}
	lock   *sync.Mutex
	wait   *sync.WaitGroup
	logger *zap.Logger
}

func NewWSClient(url string, logger *zap.Logger) *wsClient {
	ws := &wsClient{
		url:    url,
		cli:    nil,
		stop:   make(chan interface{}, 1),
		read:   make(chan interface{}, 1000),
		lock:   new(sync.Mutex),
		wait:   new(sync.WaitGroup),
		logger: logger,
	}

	return ws
}

func (c *wsClient) Ping(interval time.Duration) {
	c.wait.Add(1)
	defer c.wait.Done()

	var (
		channel = "spot.ping"
		logger  = c.logger
	)

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			if err := c.SendChannel(channel, nil); err != nil {
				logger.Error("Send.Ping", zap.Error(err))
				c.read <- errors.WithStack(err)
				return
			}
		}
	}
}

func (c *wsClient) Listen() {
	c.wait.Add(1)
	defer c.wait.Done()

	var (
		logger = c.logger
	)

	defer func() {
		if err := recover(); err != nil {
			err := fmt.Errorf("%s", err)
			c.read <- errors.WithStack(err)
		}
	}()

	for {
		select {
		case <-c.stop:
			return
		default:
			if err := c.cli.SetReadDeadline(time.Now().Add(120 * time.Second)); err != nil {
				logger.Error("SetReadDeadline", zap.Error(err))
				c.read <- errors.WithStack(err)
				return
			}

			_, msg, err := c.cli.ReadMessage()
			if err != nil {
				logger.Error("ReadMessage", zap.Error(err))
				c.read <- errors.WithStack(err)
				return
			}

			var raw struct {
				Time    int             `json:"time"`
				Channel string          `json:"channel"`
				Event   string          `json:"event"`
				Result  json.RawMessage `json:"result"`
			}

			if err := json.Unmarshal(msg, &raw); err != nil {
				logger.Error("Unmarshal", zap.Error(err))
				c.read <- errors.WithStack(err)
				return
			}

			switch raw.Event {
			case "update":
				switch raw.Channel {
				case "spot.order_book":
					var dp struct {
						T            int64                `json:"t"`
						LastUpdateId int64                `json:"lastUpdateId"`
						S            string               `json:"s"`
						Bids         [][2]decimal.Decimal `json:"bids"`
						Asks         [][2]decimal.Decimal `json:"asks"`
					}
					if err := json.Unmarshal(raw.Result, &dp); err != nil {
						c.logger.Error("Unmarshal", zap.String("data", string(raw.Result)))
						c.read <- errors.WithStack(err)
						break
					}

					c.read <- &OrderBook{
						Pair: dp.S,
						Asks: dp.Asks,
						Bids: dp.Bids,
					}
				case "spot.pong":
				}
			}
		}
	}
}

func (c *wsClient) Connect(out bool) error {
	var (
		logger = c.logger
	)

	cli, _, err := websocket.DefaultDialer.Dial(c.url+"/ws/v4/", nil)
	if err != nil {
		return err
	}

	logger.Info("connect websocket", zap.String("url", c.url))

	c.cli = cli

	go c.Ping(10 * time.Second)
	go c.Listen()

	return nil
}

func (c *wsClient) Close() error {
	close(c.stop)

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cli == nil {
		return nil
	}

	if err := c.cli.Close(); err != nil {
		if err.Error() != "tls: use of closed connection" {
			return errors.WithStack(err)
		}
	}

	c.logger.Info("关闭WS成功")

	c.cli = nil
	c.wait.Wait()
	close(c.read)

	return nil
}

func (c *wsClient) Send(msg []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cli == nil {
		return nil
	}

	if err := c.cli.WriteMessage(websocket.TextMessage, msg); err != nil {
		c.logger.Error("WriteMessage", zap.Error(err), zap.String("msg", string(msg)))
		return errors.WithStack(err)
	}
	c.logger.Info("Send", zap.String("msg", string(msg)))
	return nil
}

func (c *wsClient) SendChannel(channel string, fields map[string]interface{}) error {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["time"] = time.Now().Unix()
	fields["channel"] = channel
	msg, err := json.Marshal(fields)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Send(msg)
}

func (c *wsClient) Sub(channel string, payload []interface{}) error {
	return c.SendChannel(channel, map[string]interface{}{
		"event":   "subscribe",
		"payload": payload,
	})
}

func (c *wsClient) SubOrderBook(cp, level, interval string) error {
	channel := "spot.order_book"
	return c.Sub(channel, []interface{}{cp, level, interval})
}

func (c *wsClient) Read() interface{} {
	return <-c.read
}

func (c *wsClient) Message() chan interface{} {
	return c.read
}
