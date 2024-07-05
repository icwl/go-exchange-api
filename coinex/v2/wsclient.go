package coinex

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type WSClient struct {
	url    string
	cli    *websocket.Conn
	stop   chan interface{}
	read   chan interface{}
	lock   *sync.Mutex
	wait   *sync.WaitGroup
	logger *zap.Logger
}

func NewWSClient(url string, logger *zap.Logger) *WSClient {
	ws := &WSClient{
		url:    url,
		cli:    nil,
		stop:   nil,
		read:   nil,
		lock:   new(sync.Mutex),
		wait:   new(sync.WaitGroup),
		logger: logger,
	}

	return ws
}

func (c *WSClient) Connect() error {
	var (
		logger = c.logger
	)

	cli, _, err := websocket.DefaultDialer.Dial(c.url+"/v2/spot", nil)
	if err != nil {
		return err
	}

	logger.Info("connect websocket", zap.String("url", c.url))

	c.cli = cli
	c.stop = make(chan interface{}, 1)
	c.read = make(chan interface{}, 1000)

	go c.Ping(3 * time.Second)
	go c.Listen()

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
	close(c.read)

	return nil
}

func (c *WSClient) Ping(interval time.Duration) {
	c.wait.Add(1)
	defer c.wait.Done()

	var (
		method = "server.ping"
		params = struct{}{}
		logger = c.logger
	)

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			if err := c.SendMethod(method, params); err != nil {
				logger.Error("Send.Ping", zap.Error(err))
				c.read <- errors.WithStack(err)
				return
			}
		}
	}
}

func (c *WSClient) Listen() {
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

			msg, err = GzipDecode(msg)
			if err != nil {
				logger.Error("UnZip", zap.Error(err))
				c.read <- errors.WithStack(err)
				return
			}

			var raw struct {
				Method string          `json:"method"`
				Data   json.RawMessage `json:"data"`
				ID     interface{}     `json:"id"`
			}

			if err := json.Unmarshal(msg, &raw); err != nil {
				logger.Error("Unmarshal", zap.Error(err))
				c.read <- errors.WithStack(err)
				return
			}

			switch raw.Method {
			case "depth.update":
				var dp *SpotDepth
				if err := json.Unmarshal(raw.Data, &dp); err != nil {
					c.logger.Error("Unmarshal", zap.String("data", string(raw.Data)))
					c.read <- errors.WithStack(err)
					break
				}
				c.read <- dp
			}
		}
	}
}

func (c *WSClient) Message() chan interface{} {
	return c.read
}

func (c *WSClient) Read() interface{} {
	return <-c.read
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

func (c *WSClient) SendMethod(method string, params interface{}) error {
	msg, err := json.Marshal(map[string]interface{}{
		"id":     1,
		"method": method,
		"params": params,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Send(msg)
}

// 市场深度订阅
func (c *WSClient) SubDepth(markets []string, limit int, interval string, isFull bool) error {
	method := "depth.subscribe"
	list := make([][]interface{}, 0, len(markets))
	for _, market := range markets {
		list = append(list, []interface{}{market, limit, interval, isFull})
	}
	params := map[string]interface{}{"market_list": list}
	return c.SendMethod(method, params)
}

func GzipDecode(in []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
