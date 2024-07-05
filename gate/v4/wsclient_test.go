package gate

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWSClient_SubOrderBook(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cli := NewWSClient(wsURL, logger)
	if err := cli.Connect(); err != nil {
		t.Fatal(err)
	}
	if err := cli.SubOrderBook("BTC_USDT", "10", "100ms"); err != nil {
		t.Fatal(err)
	}
	if err := cli.SubOrderBook("ETH_USDT", "10", "100ms"); err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(60 * time.Second)
		if err := cli.Close(); err != nil {
			panic(err)
		}
	}()

	go cli.Ping(10 * time.Second)

	for {
		msg, err := cli.Read()
		if err != nil {
			t.Fatal(err)
		}
		if msg == nil {
			continue
		}
		if ob, ok := msg.(*OrderBook); ok {
			t.Logf("订单簿/深度 %s(%d,%d) 买一: %+v 卖一: %+v",
				ob.Pair,
				len(ob.Bids),
				len(ob.Asks),
				ob.Bids[0],
				ob.Asks[0])
		}
	}
}
