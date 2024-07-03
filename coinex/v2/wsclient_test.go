package coinex

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWSClient_SubDepth(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cli := NewWSClient(wsURL, logger)
	if err := cli.Connect(); err != nil {
		t.Fatal(err)
	}

	if err := cli.SubDepth([]string{"BTCUSDT"}, 10, "0", true); err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(30 * time.Second)
		if err := cli.Close(); err != nil {
			panic(err)
		}
	}()
	for msg := range cli.Message() {
		if dp, ok := msg.(*SpotDepth); ok {
			t.Logf("市场深度 (%d,%d) 买一: %+v, 卖一: %+v",
				len(dp.Depth.Asks),
				len(dp.Depth.Bids),
				dp.Depth.Bids[0],
				dp.Depth.Asks[0])
		}
	}
}
