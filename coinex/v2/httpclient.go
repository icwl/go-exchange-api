package coinex

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type httpClient struct {
	url    string
	key    string
	secret string
	cli    *http.Client
	logger *zap.Logger
}

func NewHTTPClient(url, key, secret string, logger *zap.Logger) *httpClient {
	return &httpClient{
		url:    url,
		key:    key,
		secret: secret,
		cli:    http.DefaultClient,
		logger: logger,
	}
}

func (c *httpClient) Request(method, path string, params url.Values, body map[string]interface{}, auth bool) ([]byte, error) {
	var (
		reqBody []byte
	)

	if params != nil {
		path = path + "?" + params.Encode()
	}

	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	// build request
	req, err := http.NewRequest(method, c.url+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/json")

	if auth {
		// 构建认证请求
		// set header
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		req.Header.Set("X-COINEX-KEY", c.key)
		req.Header.Set("X-COINEX-SIGN", Sign(method, path, string(reqBody), timestamp, c.secret))
		req.Header.Set("X-COINEX-TIMESTAMP", timestamp)

		c.logger.Info("Request",
			zap.String("URL", req.URL.String()),
			zap.String("Body", string(reqBody)),
			zap.Any("Header", req.Header))
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if auth {
		c.logger.Info("Response",
			zap.String("URL", req.URL.String()),
			zap.String("Body", string(respBody)))
	}

	// check status code
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Response Status",
			zap.String("URL", req.URL.String()),
			zap.String("Body", string(reqBody)),
			zap.Any("Header", req.Header),
			zap.String("Status", resp.Status),
			zap.String("Response Body", string(respBody)))
		err := NewErrResponseStatus(resp)
		return nil, errors.WithStack(err)
	}

	return respBody, nil
}

// 获取市场状态
// - market 空字符串或不传表示查询全部市场
func (c *httpClient) SpotMarket(market string) ([]*SpotMarket, error) {
	method := http.MethodGet
	path := "/v2/spot/market"
	params := url.Values{}
	if market != "" {
		params.Add("market", market)
	}

	resp, err := c.Request(method, path, params, nil, false)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int           `json:"code"`
		Data    []*SpotMarket `json:"data"`
		Message string        `json:"message"`
	}

	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := ErrResponseBody(resp)
		return nil, errors.WithStack(err)
	}

	if reply.Code != 0 {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := NewErrResponse(reply.Code, reply.Message)
		return nil, errors.WithStack(err)
	}

	return reply.Data, nil
}

// 获取市场 K 线
// - market 市场名称
// - price_type K线的绘制价格类型, 默认为latest_price, 现货市场没有mark_price
// - limit 交易数据条数. 默认 100, 最大值为 1000
// - period k 线周期. ["1min", "3min", "5min", "15min", "30min", "1hour", "2hour", "4hour", "6hour", "12hour", "1day", "3day", "1week"]中的一个
func (c *httpClient) SpotKLine(market, priceType string, limit int, period string) ([]*SpotKLine, error) {
	method := http.MethodGet
	path := "/v2/spot/kline"
	params := url.Values{}
	params.Add("market", market)
	if priceType != "" {
		params.Add("price_type", priceType)
	}
	if limit != 0 {
		params.Add("limit", strconv.Itoa(limit))
	}
	params.Add("period", period)

	resp, err := c.Request(method, path, params, nil, false)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int          `json:"code"`
		Data    []*SpotKLine `json:"data"`
		Message string       `json:"message"`
	}

	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := ErrResponseBody(resp)
		return nil, errors.WithStack(err)
	}

	if reply.Code != 0 {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := NewErrResponse(reply.Code, reply.Message)
		return nil, errors.WithStack(err)
	}

	return reply.Data, nil
}

// 获取市场深度
// - market 市场名称
// - limit 深度数据条数
// - interval 合并粒度
func (c *httpClient) SpotDepth(market string, limit int, interval string) (*SpotDepth, error) {
	method := http.MethodGet
	path := "/v2/spot/depth"
	params := url.Values{}
	params.Add("market", market)
	params.Add("limit", strconv.Itoa(limit))
	params.Add("interval", interval)

	resp, err := c.Request(method, path, params, nil, false)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int        `json:"code"`
		Data    *SpotDepth `json:"data"`
		Message string     `json:"message"`
	}

	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := ErrResponseBody(resp)
		return nil, errors.WithStack(err)
	}

	if reply.Code != 0 {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := NewErrResponse(reply.Code, reply.Message)
		return nil, errors.WithStack(err)
	}

	return reply.Data, nil
}

// 获取充提配置
func (c *httpClient) DepositWithdrawConfig(ccy string) (*DepositWithdrawConfig, error) {
	method := http.MethodGet
	path := "/v2/assets/deposit-withdraw-config"
	params := url.Values{}
	params.Add("ccy", ccy)

	resp, err := c.Request(method, path, params, nil, false)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int                    `json:"code"`
		Data    *DepositWithdrawConfig `json:"data"`
		Message string                 `json:"message"`
	}

	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := ErrResponseBody(resp)
		return nil, errors.WithStack(err)
	}

	if reply.Code != 0 {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := NewErrResponse(reply.Code, reply.Message)
		return nil, errors.WithStack(err)
	}

	return reply.Data, nil
}
