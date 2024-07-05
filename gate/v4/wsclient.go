package gate

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WSClient struct {
	url    string
	cli    *websocket.Conn
	stop   chan interface{}
	lock   *sync.Mutex
	wait   *sync.WaitGroup
	logger *zap.Logger
}

func NewWSClient(url string, logger *zap.Logger) *WSClient {
	ws := &WSClient{
		url:    url,
		cli:    nil,
		stop:   nil,
		lock:   new(sync.Mutex),
		wait:   new(sync.WaitGroup),
		logger: logger,
	}

	return ws
}

func (c *WSClient) Ping(interval time.Duration) error {
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
			return nil
		case <-ticker.C:
			if err := c.SendChannel(channel, nil); err != nil {
				logger.Error("Send.Ping", zap.Error(err))
				return errors.WithStack(err)
			}
		}
	}
}

func (c *WSClient) Connect() error {
	var (
		logger = c.logger
	)

	cli, _, err := websocket.DefaultDialer.Dial(c.url+"/ws/v4/", nil)
	if err != nil {
		return err
	}

	logger.Info("connect websocket", zap.String("url", c.url))

	c.cli = cli
	c.stop = make(chan interface{}, 1)

	return nil
}

func (c *WSClient) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cli == nil {
		return nil
	}

	close(c.stop)
	if err := c.cli.Close(); err != nil {
		if err.Error() != "tls: use of closed connection" {
			return errors.WithStack(err)
		}
	}

	c.logger.Info("关闭WS成功")

	c.cli = nil
	c.wait.Wait()

	return nil
}

func (c *WSClient) Read() (interface{}, error) {
	if err := c.cli.SetReadDeadline(time.Now().Add(120 * time.Second)); err != nil {
		return nil, errors.WithStack(err)
	}

	_, msg, err := c.cli.ReadMessage()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var raw struct {
		Time    int             `json:"time"`
		Channel string          `json:"channel"`
		Event   string          `json:"event"`
		Result  json.RawMessage `json:"result"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, errors.WithStack(err)
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
				err = errors.Wrap(err, string(raw.Result))
				return nil, errors.WithStack(err)
			}

			return &OrderBook{
				Pair: dp.S,
				Asks: dp.Asks,
				Bids: dp.Bids,
			}, nil
		case "spot.pong":
		}
	}
	return nil, nil
}

func (c *WSClient) Send(msg []byte) error {
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

func (c *WSClient) SendChannel(channel string, fields map[string]interface{}) error {
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

func (c *WSClient) Sub(channel string, payload []interface{}) error {
	return c.SendChannel(channel, map[string]interface{}{
		"event":   "subscribe",
		"payload": payload,
	})
}

func (c *WSClient) SubOrderBook(cp, level, interval string) error {
	channel := "spot.order_book"
	return c.Sub(channel, []interface{}{cp, level, interval})
}
