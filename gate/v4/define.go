package gate

import (
	"github.com/shopspring/decimal"
)

const (
	HTTPClientURL = "https://api.gateio.ws"
	WSClientURL   = "wss://api.gateio.ws"

	CandleGroupSecOneMinute      = 60     // 1分钟
	CandleGroupSecFiveMinutes    = 300    // 5分钟
	CandleGroupSecTenMinutes     = 600    // 10分钟
	CandleGroupSecFifteenMinutes = 900    // 15分钟
	CandleGroupSecTwentyMinutes  = 1200   // 20分钟
	CandleGroupSecThirtyMinutes  = 1800   // 30分钟
	CandleGroupSecOneHour        = 3600   // 1小时
	CandleGroupSecTwoHours       = 7200   // 两小时
	CandleGroupSecThreeHours     = 10800  // 3小时
	CandleGroupSecFourHours      = 14400  // 4小时
	CandleGroupSecSixHours       = 21600  // 6小时
	CandleGroupSecEightHours     = 28800  // 8小时
	CandleGroupSecTwelveHours    = 43200  // 12小时
	CandleGroupSecOneDay         = 86400  // 1天
	CandleGroupSecTwoDays        = 172800 // 2天
	CandleGroupSecOneWeek        = 604800 // 1周

	OrderTypeLimit = "limit"

	OrderStatusOpen      = "open"
	OrderStatusClosed    = "closed"
	OrderStatusCancelled = "cancelled"

	AccountSpot = "spot"
)

type Currency struct {
	// 币种名称
	Currency string `json:"currency"`
	// 是否下架
	Delisted bool `json:"delisted"`
	// 是否暂停提现
	WithdrawDisabled bool `json:"withdraw_disabled"`
	// 提现是否存在延迟
	WithdrawDelayed bool `json:"withdraw_delayed"`
	// 是否暂停充值
	DepositDisabled bool `json:"deposit_disabled"`
	// 是否暂停交易
	TradeDisabled bool `json:"trade_disabled"`
}

type CurrencyPair struct {
	// 交易对
	ID string `json:"id"`
	// 交易货币
	Base string `json:"base"`
	// 计价货币
	Quote string `json:"quote"`
	// 交易费率(要乘以百分之一)
	Fee decimal.Decimal `json:"fee"`
	// 交易货币最低交易数量，null 表示无限制
	MinBaseAmount decimal.Decimal `json:"min_base_amount,omitempty"`
	// 计价货币最低交易数量，null 表示无限制
	MinQuoteAmount decimal.Decimal `json:"min_quote_amount,omitempty"`
	// 数量精度
	AmountPrecision int32 `json:"amount_precision"`
	// 价格精度
	Precision int32 `json:"precision"`
	// untradable: 不可交易
	// buyable: 可买
	// sellable: 可卖
	// tradable: 买卖均可交易
	TradeStatus string `json:"trade_status"`
	// 允许卖出时间，秒级 Unix 时间戳
	SellStart int32 `json:"sell_start"`
	// 允许买入时间，秒级 Unix 时间戳
	BuyStart int32 `json:"buy_start"`
}

type OrderBook struct {
	Pair string
	// [[卖方价格, 卖方数量],...]
	Asks [][2]decimal.Decimal `json:"asks"`
	// [[买方价格, 买方数量],...]
	Bids [][2]decimal.Decimal `json:"bids"`
}

type Account struct {
	Currency  string          `json:"currency"`
	Available decimal.Decimal `json:"available"`
	Locked    decimal.Decimal `json:"locked"`
}

type Order struct {
	// 订单 ID
	ID string `json:"id"`
	// 订单自定义信息，用户可以用该字段设置自定义 ID，用户自定义字段必须满足以下条件：
	// 1. 必须以 t- 开头
	// 2. 不计算 t- ，长度不能超过 28 字节
	// 3. 输入内容只能包含数字、字母、下划线(_)、中划线(-) 或者点(.)
	Text string `json:"text"`
	// 订单创建时间
	CreateTime int64 `json:"create_time,string"`
	// 订单最新修改时间
	UpdateTime int64 `json:"update_time,string"`
	// 订单创建时间，毫秒精度
	CreateTimeMs int64 `json:"create_time_ms"`
	// 订单最近修改时间，毫秒精度
	UpdateTimeMs int64 `json:"update_time_ms"`
	// 订单状态。
	// open: 等待处理
	// closed: 全部成交
	// cancelled: 订单撤销
	Status string `json:"status"`
	// 交易货币对
	CurrencyPair string `json:"currency_pair"`
	// 订单类型
	// limit - 限价单
	Type string `json:"type"`
	// 账户类，
	// spot - 现货账户
	// margin - 杠杆账户
	// cross_margin - 全仓杠杆账户
	Account string `json:"account"`
	// 买单或者卖单
	Side string `json:"side"`
	// 交易数量
	Amount decimal.Decimal `json:"amount"`
	// 交易价
	Price decimal.Decimal `json:"price"`
	// Time in force 策略。
	//- gtc: GoodTillCancelled
	//- ioc: ImmediateOrCancelled，立即成交或者取消，只吃单不挂单
	//- poc: PendingOrCancelled，被动委托，只挂单不吃单
	TimeInForce string `json:"time_in_force"`
	// 冰山下单显示的数量，不指定或传 0 都默认为普通下单。如果需要全部冰山，设置为 -1
	Iceberg decimal.Decimal `json:"iceberg"`
	// 交易货币未成交数量
	Left decimal.Decimal `json:"left"`
	// 已成交的计价币种总额，该字段废弃，建议使用相同意义的 filled_total
	FillPrice decimal.Decimal `json:"fill_price"`
	// 已成交总金额
	FilledTotal decimal.Decimal `json:"filled_total"`
	// 成交扣除的手续费
	Fee decimal.Decimal `json:"fee"`
	// 手续费计价单位
	FeeCurrency string `json:"fee_currency"`
	// 手续费抵扣使用的点卡数量
	PointFee decimal.Decimal `json:"point_fee"`
	// 手续费抵扣使用的 GT 数量
	GtFee decimal.Decimal `json:"gt_fee"`
	// 是否开启GT抵扣
	GtDiscount bool `json:"gt_discount"`
	// 返还的手续费
	RebatedFee decimal.Decimal `json:"rebated_fee"`
	// 返还手续费计价单位
	RebatedFeeCurrency string `json:"rebated_fee_currency"`
}

type Address struct {
	Currency            string `json:"currency"`
	Address             string `json:"address"`
	MultiChainAddresses []struct {
		Chain        string `json:"chain"`
		Address      string `json:"address"`
		PaymentId    string `json:"payment_id"`
		PaymentName  string `json:"payment_name"`
		ObtainFailed int    `json:"obtain_failed"`
	} `json:"multichain_addresses"`
}

type Withdrawal struct {
	ID        string          `json:"id"`
	Timestamp int64           `json:"timestamp,string"`
	Currency  string          `json:"currency"`
	Address   string          `json:"address"`
	TxId      string          `json:"txid"`
	Amount    decimal.Decimal `json:"amount"`
	Memo      string          `json:"memo"`
	Status    string          `json:"status"`
	Chain     string          `json:"chain"`
}
