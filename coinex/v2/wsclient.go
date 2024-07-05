package coinex

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

func (c *WSClient) Ping(interval time.Duration) error {
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
			return nil
		case <-ticker.C:
			if err := c.SendMethod(method, params); err != nil {
				logger.Error("Send.Ping", zap.Error(err))
				return errors.WithStack(err)
			}
		}
	}
}

func (c *WSClient) Read() (interface{}, error) {
	if err := c.cli.SetReadDeadline(time.Now().Add(120 * time.Second)); err != nil {
		return nil, errors.WithStack(err)
	}

	_, msg, err := c.cli.ReadMessage()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	msg, err = GzipDecode(msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var raw struct {
		Method string          `json:"method"`
		Data   json.RawMessage `json:"data"`
		ID     interface{}     `json:"id"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, errors.WithStack(err)
	}

	switch raw.Method {
	case "depth.update":
		var dp *SpotDepth
		if err := json.Unmarshal(raw.Data, &dp); err != nil {
			err = errors.Wrap(err, string(raw.Data))
			return nil, errors.WithStack(err)
		}
		return dp, nil
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
