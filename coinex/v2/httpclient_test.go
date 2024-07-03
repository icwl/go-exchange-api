package coinex

import (
	"flag"
	"testing"

	"go.uber.org/zap"
)

var (
	pKey     = flag.String("key", "", "key")
	pSecret  = flag.String("secret", "", "secret")
	pHTTPURL = flag.String("http_url", HTTPURL, "http_url")
	pWSURL   = flag.String("ws_url", WSURL, "ws_url")

	key     string
	secret  string
	httpURL string
	wsURL   string
)

func init() {
	testing.Init()

	flag.Parse()

	key = *pKey
	secret = *pSecret
	httpURL = *pHTTPURL
	wsURL = *pWSURL
}

func TestHTTPClient_SpotMarket(t *testing.T) {

	logger := zap.NewExample()
	client := NewHTTPClient(httpURL, key, secret, logger)

	res, err := client.SpotMarket("")
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range res {
		t.Logf("市场状态 : %+v", item)
	}
}

func TestHTTPClient_SpotKLine(t *testing.T) {

	logger := zap.NewExample()
	client := NewHTTPClient(httpURL, key, secret, logger)

	res, err := client.SpotKLine("LATUSDT", "", 10, "15min")
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range res {
		t.Logf("K 线 : %+v", item)
	}
}

func TestHTTPClient_SpotDepth(t *testing.T) {

	logger := zap.NewExample()
	client := NewHTTPClient(httpURL, key, secret, logger)

	res, err := client.SpotDepth("BTCUSDT", 5, "0")
	if err != nil {
		t.Fatal(err)
	}

	for i := len(res.Depth.Asks) - 1; i >= 0; i-- {
		t.Logf("卖方报价 %d : %+v", i+1, res.Depth.Asks[i])
	}
	t.Log("----------------")
	for i, item := range res.Depth.Bids {
		t.Logf("买方报价 %d : %+v", i+1, item)
	}
}

func TestHTTPClient_DepositWithdrawConfig(t *testing.T) {
	logger := zap.NewExample()
	client := NewHTTPClient(httpURL, key, secret, logger)

	res, err := client.DepositWithdrawConfig("USDT")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("获取充提配置 : %+v", res)
}
