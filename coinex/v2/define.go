package coinex

import (
	"github.com/shopspring/decimal"
)

const (
	HTTPURL = "https://api.coinex.com"
	WSURL   = "wss://socket.coinex.com"

	WithdrawMethodOnChain   = "on_chain"
	WithdrawMethodInterUser = "inter_user"

	MarketTypeSpot    = "SPOT"
	MarketTypeMargin  = "MARGIN"
	MarketTypeFutures = "FUTURES"

	OrderTypeLimit     = "limit"      // 限价单, 一直生效, GTC 订单
	OrderTypeMarket    = "market"     // 市价单
	OrderTypeMakerOnly = "maker_only" // 只做 maker 单, post_only 订单
	OrderTypeIOC       = "ioc"        // 立即成交或取消
	OrderTypeFOK       = "fok"        // 全部成交或全部取消

	OrderStatusOpen         = "open"          // 已提交，等待成交;
	OrderStatusPartFilled   = "part_filled"   // 部分成交(订单仍在挂单中);
	OrderStatusFilled       = "filled"        // 完全成交(订单已完成);
	OrderStatusPartCanceled = "part_canceled" // 已撤销部分成交(订单成交部分后被撤销);
	OrderStatusCanceled     = "canceled"      // 订单已取消(订单已完成);为了保证服务器性能，所有取消的没有任何成交的订单均不会保存
)

type SpotMarket struct {
	// 市场名称
	Market string `json:"market"`
	// maker手续费费率
	MakerFeeRate decimal.Decimal `json:"maker_fee_rate"`
	// taker手续费费率
	TakerFeeRate decimal.Decimal `json:"taker_fee_rate"`
	// 最小交易量
	MinAmount decimal.Decimal `json:"min_amount"`
	// 交易币种
	BaseCcy string `json:"base_ccy"`
	// 报价币种
	QuoteCcy string `json:"quote_ccy"`
	// 交易币种精度
	BaseCcyPrecision int32 `json:"base_ccy_precision"`
	// 报价币种精度
	QuoteCcyPrecision int32 `json:"quote_ccy_precision"`
	IsAmmAvailable    bool  `json:"is_amm_available"`
	IsMarginAvailable bool  `json:"is_margin_available"`
}

type SpotKLine struct {
	// 市场名称
	Market string `json:"market"`
	// 时间戳
	CreatedAt int64 `json:"created_at"`
	// 开盘价
	Open decimal.Decimal `json:"open"`
	// 收盘价
	Close decimal.Decimal `json:"close"`
	// 最高价
	High decimal.Decimal `json:"high"`
	// 最低价
	Low decimal.Decimal `json:"low"`
	// 成交量
	Volume decimal.Decimal `json:"volume"`
	// 成交额
	Value decimal.Decimal `json:"value"`
}

type SpotDepth struct {
	Depth struct {
		// [[卖方价格, 卖方数量],...]
		Asks [][2]decimal.Decimal `json:"asks"`
		// [[买方价格, 买方数量],...]
		Bids     [][2]decimal.Decimal `json:"bids"`
		Checksum int64                `json:"checksum"`
		// 最新价格
		Last      decimal.Decimal `json:"last"`
		UpdatedAt int64           `json:"updated_at"`
	} `json:"depth"`
	// true为全量推送, false为增量推送
	IsFull bool `json:"is_full"`
	// 市场名称
	Market string `json:"market"`
}

type DepositWithdrawConfig struct {
	Asset struct {
		// 币种名称
		Ccy            string `json:"ccy"`
		DepositEnabled bool   `json:"deposit_enabled"`
		// 是否开启充值
		WithdrawEnabled bool `json:"withdraw_enabled"`
		// 是否开启提现
		InterTransferEnabled bool `json:"inter_transfer_enabled"`
		IsSt                 bool `json:"is_st"`
	} `json:"asset"`
	Chains []struct {
		// 公链名称
		Chain string `json:"chain"`
		// 最小充值量
		MinDepositAmount decimal.Decimal `json:"min_deposit_amount"`
		// 最小提现量
		MinWithdrawAmount decimal.Decimal `json:"min_withdraw_amount"`
		// 是否开启充值
		DepositEnabled bool `json:"deposit_enabled"`
		// 是否开启提现
		WithdrawEnabled           bool   `json:"withdraw_enabled"`
		DepositDelayMinutes       int    `json:"deposit_delay_minutes"`
		SafeConfirmations         int    `json:"safe_confirmations"`
		IrreversibleConfirmations int    `json:"irreversible_confirmations"`
		DeflationRate             string `json:"deflation_rate"`
		// 提现手续费
		WithdrawalFee decimal.Decimal `json:"withdrawal_fee"`
		// 提现精度
		WithdrawalPrecision      int    `json:"withdrawal_precision"`
		Memo                     string `json:"memo"`
		IsMemoRequiredForDeposit bool   `json:"is_memo_required_for_deposit"`
		ExplorerAssetURL         string `json:"explorer_asset_url"`
	} `json:"chains"`
}

type DepositAddress struct {
	Address string `json:"address"`
	Memo    string `json:"memo"`
}

type Withdraw struct {
	WithdrawID         int64           `json:"withdraw_id"`
	CreatedAt          int64           `json:"created_at"`
	Ccy                string          `json:"ccy"`
	Chain              string          `json:"chain"`
	Amount             decimal.Decimal `json:"amount"`
	ActualAmount       decimal.Decimal `json:"actual_amount"`
	WithdrawMethod     string          `json:"withdraw_method"`
	Memo               string          `json:"memo"`
	TxFee              decimal.Decimal `json:"tx_fee"`
	TxID               string          `json:"tx_id"`
	ToAddress          string          `json:"to_address"`
	Confirmations      int             `json:"confirmations"`
	ExplorerAddressURL string          `json:"explorer_address_url"`
	ExplorerTxURL      string          `json:"explorer_tx_url"`
	Status             string          `json:"status"`
	Remark             string          `json:"remark"`
}

type SpotBalance struct {
	Ccy       string          `json:"ccy"`
	Available decimal.Decimal `json:"available"`
	Frozen    decimal.Decimal `json:"frozen"`
}

type SpotOrder struct {
	// 订单 ID
	OrderID int64 `json:"order_id"`
	// 市场名称
	Market string `json:"market"`
	// 市场类型
	MarketType string `json:"market_type"`
	// 币种名称
	Ccy string `json:"ccy"`
	// 订单方向
	Side string `json:"side"`
	// 订单类型
	Type string `json:"type"`
	// 委托数量
	Amount decimal.Decimal `json:"amount"`
	// 委托价格
	Price decimal.Decimal `json:"price"`
	// 未成交数量
	UnfilledAmount decimal.Decimal `json:"unfilled_amount"`
	// 已成交数量
	FilledAmount decimal.Decimal `json:"filled_amount"`
	// 已成交价值
	FilledValue decimal.Decimal `json:"filled_value"`
	// 客户端 ID
	ClientID string `json:"client_id"`
	// 收取的交易币种手续费
	BaseFee decimal.Decimal `json:"base_fee"`
	// 收取的报价币种手续费
	QuoteFee decimal.Decimal `json:"quote_fee"`
	// 收取的抵扣币种手续费
	DiscountFee decimal.Decimal `json:"discount_fee"`
	// maker 手续费费率
	MakerFeeRate decimal.Decimal `json:"maker_fee_rate"`
	// taker 手续费费率
	TakerFeeRate decimal.Decimal `json:"taker_fee_rate"`
	CreatedAt    int64           `json:"created_at"`
	UpdatedAt    int64           `json:"updated_at"`
	// 注意: last_fill_amount,last_fill_price,status可能不返回
	LastFillAmount string `json:"last_fill_amount"`
	LastFillPrice  string `json:"last_fill_price"`
	Status         string `json:"status"`
}
