package gate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type HTTPClient interface {
	// 查询所有币种信息
	Currencies() ([]*Currency, error)
	// 查询支持的所有交易对
	CurrencyPairs() ([]*CurrencyPair, error)
	// 获取市场深度信息
	OrderBook(pair, interval string, limit int) (*OrderBook, error)
	// 获取现货交易账户列表
	Accounts(currency string) ([]*Account, error)
	// 查询所有挂单
	OpenOrders(page, limit int, account string) ([]*Order, error)
	NewOrder(text, pair, type_, account, side, amount, price string) (*Order, error)
	CancelOrder(orderId, pair, account string) (*Order, error)
	GetOrder(orderId, pair, account string) (*Order, error)
	DepositAddress(currency string) (*Address, error)
	Withdrawal(amount, currency, address, memo, chain string) (*Withdrawal, error)
}

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

func (c *httpClient) Request(method, path string, query url.Values, body map[string]interface{}, auth bool) ([]byte, error) {
	var (
		rawQuery = query.Encode()
		reqBody  []byte
	)

	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	req, err := http.NewRequest(method, c.url+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.URL.RawQuery = rawQuery
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if auth {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Add("KEY", c.key)
		req.Header.Add("SIGN", Sign(method, path, rawQuery, reqBody, timestamp, c.secret))
		req.Header.Add("Timestamp", timestamp)

		c.logger.Info("Request",
			zap.String("URL", req.URL.String()),
			zap.String("Body", string(reqBody)),
			zap.Any("Header", req.Header))
	}

	// 发出请求
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	// 解析响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if auth {
		c.logger.Info("Response",
			zap.String("URL", req.URL.String()),
			zap.String("Body", string(respBody)))
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated {
		c.logger.Error("Response Status",
			zap.String("URL", req.URL.String()),
			zap.String("Body", string(reqBody)),
			zap.Any("Header", req.Header),
			zap.String("Status", resp.Status),
			zap.String("Response Body", string(respBody)))
		err := NewErrResponse(respBody)
		return nil, errors.WithStack(err)
	}

	return respBody, nil
}

// 查询所有币种信息
func (c *httpClient) Currencies() ([]*Currency, error) {
	method := http.MethodGet
	path := "/api/v4/spot/currencies"
	respBody, err := c.Request(method, path, nil, nil, false)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply []*Currency
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}

// 查询支持的所有交易对
func (c *httpClient) CurrencyPairs() ([]*CurrencyPair, error) {
	method := http.MethodGet
	path := "/api/v4/spot/currency_pairs"
	respBody, err := c.Request(method, path, nil, nil, false)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply []*CurrencyPair
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}

// 获取市场深度信息
func (c *httpClient) OrderBook(pair, interval string, limit int) (*OrderBook, error) {
	method := http.MethodGet
	path := "/api/v4/spot/order_book"
	query := url.Values{}
	query.Add("currency_pair", pair)
	if interval != "" {
		query.Add("interval", interval)
	}
	if limit != 0 {
		query.Add("limit", strconv.Itoa(limit))
	}
	respBody, err := c.Request(method, path, query, nil, false)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}
	var reply struct {
		Current int64 `json:"current"`
		Update  int64 `json:"update"`
		// 卖方深度
		Asks [][2]decimal.Decimal `json:"asks"`
		// 买方深度
		Bids [][2]decimal.Decimal `json:"bids"`
	}
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return &OrderBook{
		Pair: pair,
		Asks: reply.Asks,
		Bids: reply.Bids,
	}, nil
}

// 获取现货交易账户列表
func (c *httpClient) Accounts(currency string) ([]*Account, error) {
	method := http.MethodGet
	path := "/api/v4/spot/accounts"
	query := url.Values{}
	if currency != "" {
		query.Add("currency", currency)
	}
	respBody, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply []*Account
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}

// 查询所有挂单
func (c *httpClient) OpenOrders(page, limit int, account string) ([]*Order, error) {
	method := http.MethodGet
	path := "/api/v4/spot/open_orders"
	query := url.Values{}
	if page != 0 {
		query.Add("page", strconv.Itoa(page))
	}
	if limit != 0 {
		query.Add("limit", strconv.Itoa(limit))
	}
	if account != "" {
		query.Add("account", account)
	}
	respBody, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply []struct {
		CurrencyPair string   `json:"currency_pair"`
		Total        int      `json:"total"`
		Orders       []*Order `json:"orders"`
	}
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	orders := make([]*Order, 0, len(reply))
	for _, item := range reply {
		orders = append(orders, item.Orders...)
	}
	return orders, nil
}

func (c *httpClient) NewOrder(text, pair, type_, account, side, amount, price string) (*Order, error) {
	method := http.MethodPost
	path := "/api/v4/spot/orders"
	body := make(map[string]interface{})
	if text != "" {
		body["text"] = text
	}
	body["currency_pair"] = pair
	if type_ != "" {
		body["type"] = type_
	}
	if account != "" {
		body["account"] = account
	}
	body["side"] = side
	body["amount"] = amount
	body["price"] = price

	respBody, err := c.Request(method, path, nil, body, true)
	if err != nil {
		c.logger.Error(method+" "+path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply *Order
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(method+" "+path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}

func (c *httpClient) CancelOrder(orderId, pair, account string) (*Order, error) {
	method := http.MethodDelete
	path := fmt.Sprintf("/api/v4/spot/orders/%s", orderId)
	query := url.Values{}
	query.Add("currency_pair", pair)
	if account != "" {
		query.Add("account", account)
	}

	respBody, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(method+" "+path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply *Order
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(method+" "+path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}

func (c *httpClient) GetOrder(orderId, pair, account string) (*Order, error) {
	method := http.MethodGet
	path := fmt.Sprintf("/api/v4/spot/orders/%s", orderId)
	query := url.Values{}
	query.Add("currency_pair", pair)
	if account != "" {
		query.Add("account", account)
	}
	respBody, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply *Order
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}

func (c *httpClient) DepositAddress(currency string) (*Address, error) {
	method := http.MethodGet
	path := "/api/v4/wallet/deposit_address"
	query := url.Values{}
	query.Add("currency", currency)

	respBody, err := c.Request(method, path, query, nil, true)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply *Address
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}

func (c *httpClient) Withdrawal(amount, currency, address, memo, chain string) (*Withdrawal, error) {
	method := http.MethodPost
	path := "/api/v4/withdrawals"
	body := make(map[string]interface{})
	body["amount"] = amount
	body["currency"] = currency
	body["address"] = address
	if memo != "" {
		body["memo"] = memo
	}
	if chain != "" {
		body["chain"] = chain
	}

	respBody, err := c.Request(method, path, nil, body, true)
	if err != nil {
		c.logger.Error(method+" "+path, zap.Error(err))
		return nil, errors.WithStack(err)
	}

	var reply *Withdrawal
	if err := json.Unmarshal(respBody, &reply); err != nil {
		c.logger.Error(method+" "+path, zap.String("reply", string(respBody)), zap.Error(err))
		err := ErrResponseBody(respBody)
		return nil, errors.WithStack(err)
	}

	return reply, nil
}
