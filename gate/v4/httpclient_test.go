package gate

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

var (
	pKey     = flag.String("key", "", "key")
	pSecret  = flag.String("secret", "", "secret")
	pSymbol  = flag.String("symbol", "btc_usdt", "symbol")
	pHTTPURL = flag.String("http_url", HTTPClientURL, "http_url")
	pWSURL   = flag.String("ws_url", WSClientURL, "ws_url")

	key     string
	secret  string
	symbol  string
	httpURL string
	wsURL   string
)

func init() {
	testing.Init()

	flag.Parse()

	key = *pKey
	secret = *pSecret
	symbol = *pSymbol
	httpURL = *pHTTPURL
	wsURL = *pWSURL
}

func TestHTTPClient_Currencies(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)
	res, err := cli.Currencies()
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range res {
		if item.Currency != "NUM" {
			continue
		}
		t.Logf("Currency : %+v", item)
	}
}

func TestHTTPClient_CurrencyPairs(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)
	res, err := cli.CurrencyPairs()
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range res {
		t.Logf("Pair : %v", item)
	}
}

func TestHTTPClient_OrderBook(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)
	dp, err := cli.OrderBook(symbol, "", 10)
	if err != nil {
		t.Fatal(err)
	}

	for _, a := range dp.Asks {
		t.Logf("Depth : Asks: %v", a)
	}

	t.Log("----------------------------------")
	for _, b := range dp.Bids {
		t.Logf("Depth : Bids: %v", b)
	}
}

func TestHTTPClient_Accounts(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)
	lis, err := cli.Accounts("")
	if err != nil {
		t.Fatal(err)
	}

	for _, b := range lis {
		if b.Available.Add(b.Locked).IsZero() {
			continue
		}
		t.Logf("Balances : %v", b)
	}
}

func TestHTTPClient_OpenOrders(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)

	res, err := cli.OpenOrders(0, 0, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range res {
		t.Logf("Order : %v", item)
	}
}

func TestHTTPClient_NewOrder(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)
	var (
		pair    = "BTC_USDT"
		type_   = "limit"
		account = "spot"
		side    = "buy"
		amount  = "0.1"
		price   = "100"
		text    = fmt.Sprintf("t-%s_%s_%d", symbol, side, time.Now().Unix())
	)
	order, err := cli.NewOrder(text, pair, type_, account, side, amount, price)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Order : %v", order)
}

func TestHTTPClient_CancelOrder(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)

	//order, err := cli.CancelOrder("106358528112", "BTC_USDT", "")
	order, err := cli.CancelOrder("107266744517", "BTC_USDT", "")
	//order, err := cli.CancelOrder("106365865817", "GOLD_USDT", "")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Order : %v", order)
}

func TestHTTPClient_GetOrder(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)

	order, err := cli.GetOrder("106358528112", "BTC_USDT", "")
	//order, err := cli.GetOrder("106366474094", "GOLD_USDT", "")
	//order, err := cli.GetOrder("106365865817", "GOLD_USDT", "")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Order : %v", order)
}

func TestHTTPClient_DepositAddress(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)

	address, err := cli.DepositAddress("XAVA")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Address : %+v", address.Address)
	for _, item := range address.MultiChainAddresses {
		t.Logf("Address : %+v", item)
	}
}

func TestHTTPClient_Withdrawal(t *testing.T) {
	logger := zap.NewExample()
	cli := NewHTTPClient(httpURL, key, secret, logger)

	var (
		amount   = "3.60199268"
		currency = "NUM"
		address  = "0xc9749553bdce6daa08dfd5802222a5fc9f264844"
		memo     = ""
		chain    = "BSC"
	)
	withdrawal, err := cli.Withdrawal(amount, currency, address, memo, chain)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Withdrawal : %+v", withdrawal)
}
