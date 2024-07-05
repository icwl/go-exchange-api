package coinex

import (
	"github.com/shopspring/decimal"
)

const (
	HTTPURL = "https://api.coinex.com"
	WSURL   = "wss://socket.coinex.com"
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

type SpotBalance struct {
	Ccy       string          `json:"ccy"`
	Available decimal.Decimal `json:"available"`
	Frozen    decimal.Decimal `json:"frozen"`
}
