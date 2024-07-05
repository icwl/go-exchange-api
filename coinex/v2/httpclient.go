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

type HTTPClient struct {
	url    string
	key    string
	secret string
	cli    *http.Client
	logger *zap.Logger
}

func NewHTTPClient(url, key, secret string, logger *zap.Logger) *HTTPClient {
	return &HTTPClient{
		url:    url,
		key:    key,
		secret: secret,
		cli:    http.DefaultClient,
		logger: logger,
	}
}

func (c *HTTPClient) Request(method, path string, query url.Values, body map[string]interface{}, auth bool) ([]byte, error) {
	var (
		reqBody []byte
	)

	if query != nil {
		path = path + "?" + query.Encode()
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
func (c *HTTPClient) SpotMarket(market string) ([]*SpotMarket, error) {
	method := http.MethodGet
	path := "/v2/spot/market"
	query := url.Values{}
	if market != "" {
		query.Add("market", market)
	}

	resp, err := c.Request(method, path, query, nil, false)
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
func (c *HTTPClient) SpotKLine(market, priceType string, limit int, period string) ([]*SpotKLine, error) {
	method := http.MethodGet
	path := "/v2/spot/kline"
	query := url.Values{}
	query.Add("market", market)
	if priceType != "" {
		query.Add("price_type", priceType)
	}
	if limit != 0 {
		query.Add("limit", strconv.Itoa(limit))
	}
	query.Add("period", period)

	resp, err := c.Request(method, path, query, nil, false)
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
func (c *HTTPClient) SpotDepth(market string, limit int, interval string) (*SpotDepth, error) {
	method := http.MethodGet
	path := "/v2/spot/depth"
	query := url.Values{}
	query.Add("market", market)
	query.Add("limit", strconv.Itoa(limit))
	query.Add("interval", interval)

	resp, err := c.Request(method, path, query, nil, false)
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
func (c *HTTPClient) DepositWithdrawConfig(ccy string) (*DepositWithdrawConfig, error) {
	method := http.MethodGet
	path := "/v2/assets/deposit-withdraw-config"
	query := url.Values{}
	query.Add("ccy", ccy)

	resp, err := c.Request(method, path, query, nil, false)
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

func (c *HTTPClient) DepositAddress(ccy, chain string) (*DepositAddress, error) {
	method := http.MethodGet
	path := "/v2/assets/deposit-address"
	query := url.Values{}
	query.Add("ccy", ccy)
	query.Add("chain", chain)

	resp, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int             `json:"code"`
		Data    *DepositAddress `json:"data"`
		Message string          `json:"message"`
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

// - ccy 币种名称
// - chain 公链名称, 链上提现时必填, 站内转账时不必填
// - to_address 提现地址. 提现地址需要先通过开发者控制台添加IP白名单, 链上提现或者站内转账都需要先添加白名单
// - withdraw_method 提现方法, 链上提现(on_chain) 或者 站内转账(inter_user), 默认为链上提现
// - memo 部分币种充值需要 memo
// - amount 提现数量
// - extra 如果是KDA链的提现, 需要在extra字段中附加chain_id字段
// - remark 提现备注
func (c *HTTPClient) Withdraw(ccy, chain, toAddress, withdrawMethod, memo, amount string, extra interface{}, remark string) (*Withdraw, error) {
	method := http.MethodPost
	path := "/v2/assets/withdraw"
	body := make(map[string]interface{})
	body["ccy"] = ccy
	if chain != "" {
		body["chain"] = chain
	}
	body["to_address"] = toAddress
	if withdrawMethod != "" {
		body["withdraw_method"] = withdrawMethod
	}
	if memo != "" {
		body["memo"] = memo
	}
	body["amount"] = amount
	if extra != nil {
		body["extra"] = extra
	}
	if remark != "" {
		body["remark"] = remark
	}

	resp, err := c.Request(method, path, nil, body, true)
	if err != nil {
		c.logger.Error(method+" "+path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int       `json:"code"`
		Data    *Withdraw `json:"data"`
		Message string    `json:"message"`
	}
	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(method+" "+path, zap.String("resp", string(resp)), zap.Error(err))
		err := ErrResponseBody(resp)
		return nil, errors.WithStack(err)
	}

	if reply.Code != 0 {
		c.logger.Error(method+" "+path, zap.String("resp", string(resp)), zap.Error(err))
		err := NewErrResponse(reply.Code, reply.Message)
		return nil, errors.WithStack(err)
	}

	return reply.Data, nil
}

func (c *HTTPClient) SpotBalance() ([]*SpotBalance, error) {
	method := http.MethodGet
	path := "/v2/assets/spot/balance"

	resp, err := c.Request(method, path, nil, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int            `json:"code"`
		Data    []*SpotBalance `json:"data"`
		Message string         `json:"message"`
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

// - market 市场名称
// - market_type 市场类型 [SPOT / MARGIN / FUTURES]
// - side 订单方向 [buy / sell]
// - type 订单类型 [limit / market / maker_only / ioc / fok]
// - amount 委托数量
// - price 委托价格
// - client_id 客户自定义 ID
func (c *HTTPClient) SpotOrder(market, marketType, side, type_, ccy, amount, price, clientId string) (*SpotOrder, error) {
	method := http.MethodPost
	path := "/v2/spot/order"
	body := make(map[string]interface{})
	body["market"] = market
	body["market_type"] = marketType
	body["side"] = side
	body["type"] = type_
	if ccy != "" {
		body["ccy"] = ccy
	}
	body["amount"] = amount
	if price != "" {
		body["price"] = price
	}
	if clientId != "" {
		body["client_id"] = clientId
	}

	resp, err := c.Request(method, path, nil, body, true)
	if err != nil {
		c.logger.Error(method+" "+path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int        `json:"code"`
		Data    *SpotOrder `json:"data"`
		Message string     `json:"message"`
	}
	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(method+" "+path, zap.String("resp", string(resp)), zap.Error(err))
		err := ErrResponseBody(resp)
		return nil, errors.WithStack(err)
	}

	if reply.Code != 0 {
		c.logger.Error(method+" "+path, zap.String("resp", string(resp)), zap.Error(err))
		err := NewErrResponse(reply.Code, reply.Message)
		return nil, errors.WithStack(err)
	}

	return reply.Data, nil
}

func (c *HTTPClient) SpotCancelOrder(market, marketType string, orderID int64) (*SpotOrder, error) {
	method := http.MethodPost
	path := "/v2/spot/cancel-order"
	body := make(map[string]interface{})
	body["market"] = market
	body["market_type"] = marketType
	body["order_id"] = orderID

	resp, err := c.Request(method, path, nil, body, true)
	if err != nil {
		c.logger.Error(method+" "+path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int        `json:"code"`
		Data    *SpotOrder `json:"data"`
		Message string     `json:"message"`
	}
	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(method+" "+path, zap.String("resp", string(resp)), zap.Error(err))
		err := ErrResponseBody(resp)
		return nil, errors.WithStack(err)
	}

	if reply.Code != 0 {
		c.logger.Error(method+" "+path, zap.String("resp", string(resp)), zap.Error(err))
		err := NewErrResponse(reply.Code, reply.Message)
		return nil, errors.WithStack(err)
	}

	return reply.Data, nil
}

func (c *HTTPClient) SpotOrderStatus(market string, orderID int64) (*SpotOrder, error) {
	method := http.MethodGet
	path := "/v2/spot/order-status"
	query := url.Values{}
	query.Add("market", market)
	query.Add("order_id", strconv.FormatInt(orderID, 10))

	resp, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code    int        `json:"code"`
		Data    *SpotOrder `json:"data"`
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

func (c *HTTPClient) SpotFinishedOrder(market, market_type, side string, page, limit int) ([]*SpotOrder, error) {
	method := http.MethodGet
	path := "/v2/spot/finished-order"
	query := url.Values{}
	if market != "" {
		query.Add("market", market)
	}
	query.Add("market_type", market_type)
	if side != "" {
		query.Add("side", side)
	}
	if page != 0 {
		query.Add("page", strconv.Itoa(page))
	}
	if limit != 0 {
		query.Add("limit", strconv.Itoa(limit))
	}

	resp, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply struct {
		Code       int          `json:"code"`
		Data       []*SpotOrder `json:"data"`
		Message    string       `json:"message"`
		Pagination struct {
			HasNext bool `json:"has_next"`
		} `json:"pagination"`
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
