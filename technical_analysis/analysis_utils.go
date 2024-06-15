package technical_analysis

import (
	"math"
	"sort"

	"github.com/hsuloong/stock_speculation/data_center"
)

// @func 计算历史价格分位
// @return 0-100之间 也就是原始值乘了100
func CalculatePeriodRelativelyPercent(kline_items []data_center.KlineItem, start int, end int, target int) float64 {
	array_len := end - start
	close_price := make([]int64, array_len)
	for i := start; i < end; i++ {
		close_price[i-start] = kline_items[i].Close
	}

	sort.Slice(close_price, func(i, j int) bool {
		return close_price[i] < close_price[j]
	})

	price_percent := 0.0
	for i := 0; i < array_len; i++ {
		price_percent = float64(i) / float64(array_len)
		if kline_items[target].Close < close_price[i] {
			break
		}
	}

	return price_percent * 100.0
}

// @func 计算历史最低价需要下跌的百分比
// @return 0-100之间 也就是原始值乘了100
func CalculatePeriodToLowestPercent(kline_items []data_center.KlineItem, start int, end int, target int) float64 {
	min_close_price := int64(math.MaxInt64)
	for i := start; i < end; i++ {
		if min_close_price > kline_items[i].Close {
			min_close_price = kline_items[i].Close
		}
	}

	return (1.0 - (float64(min_close_price) / float64(kline_items[target].Close))) * 100
}

// @func 计算波动率
func CalculatePeriodVolatility(kline_items []data_center.KlineItem, start int, end int) float64 {
	percent_sum := 0.0
	for i := start; i < end; i++ {
		percent_sum += kline_items[i].Percent
	}
	n := float64(end - start)

	avg_percent := percent_sum / n

	std_percent_sum := 0.0
	for i := start; i < end; i++ {
		std_percent_sum += (kline_items[i].Percent - avg_percent) * (kline_items[i].Percent - avg_percent)
	}

	volatility := math.Pow(std_percent_sum/(n-1.0), 0.5)

	return volatility
}

// @func 计算周期内平均成交金额
func CalculatePeriodAvgAmount(kline_items []data_center.KlineItem, start int, end int) float64 {
	amount_sum := 0.0
	for i := start; i < end; i++ {
		amount_sum += float64(kline_items[i].Amount)
	}

	return amount_sum / float64(end-start)
}

// @func 计算周期内网格交易收益
// @return 原始值乘了100
func CalculateGridTradingProfit(kline_items []data_center.KlineItem, start int, end int, grid_percent float64) float64 {
	const kMaxSharesPerTrade int64 = 100000 // 单次交易股数
	const kTax float64 = 3.0 / 1000.0       // 交易费率 简单处理下

	var grid_price int64 = 0
	net_buy_count := 0
	account_money := make([]float64, end-start+1) // 记账

	for i := start; i < end; i++ {
		// 首次建仓直接买入
		if i == start {
			grid_price = kline_items[i].Open
			account_money[i-start] -= float64(grid_price*kMaxSharesPerTrade) * (1 + kTax)
			net_buy_count++
			continue
		}
		next_buy_price := int64(float64(grid_price) * (1.0 - grid_percent))
		next_sell_price := int64(float64(grid_price) * (1.0 + grid_percent))

		if next_buy_price >= kline_items[i].EntityLow { // 买优先
			grid_price = next_buy_price
			account_money[i-start] -= float64(grid_price*kMaxSharesPerTrade) * (1 + kTax)
			net_buy_count++
		} else if next_sell_price <= kline_items[i].EntityHigh {
			if net_buy_count > 0 {
				grid_price = next_sell_price
				account_money[i-start] += float64(grid_price*kMaxSharesPerTrade) * (1 - kTax)
				net_buy_count--
			} else {
				grid_price = kline_items[i].Close // 更新网格 不更新容易买不到了
			}
		}
	}

	// 未卖出的以当日实体最高价卖出
	if net_buy_count > 0 {
		account_money[end-start-1] += float64(kline_items[end-1].EntityHigh*kMaxSharesPerTrade) * (1 - kTax) * float64(net_buy_count)
	}

	invest_money := -(CalculateMinSubArraySum(account_money))
	profit := CalculateArraySum(account_money)
	if invest_money > 1e-3 {
		return profit / invest_money * 100.0
	} else {
		return 0.0
	}
}

// @func 计算最小连续数组和
func CalculateMinSubArraySum(array []float64) float64 {
	min_sum := array[0]
	last_min_sum := array[0]

	for i := 1; i < len(array); i++ {
		if last_min_sum <= 0 {
			last_min_sum += array[i]
		} else {
			last_min_sum = array[i]
		}

		min_sum = math.Min(min_sum, last_min_sum)
	}

	return min_sum
}

// @func 计算数组和
func CalculateArraySum(array []float64) float64 {
	array_sum := 0.0
	for i := 0; i < len(array); i++ {
		array_sum += array[i]
	}

	return array_sum
}

// @func 计算周期内涨停次数
func CalculateStockLimitUpTimes(kline_items []data_center.KlineItem, start int, end int, limit_up_rate float64) int64 {
	var result int64 = 0
	for i := start; i < end; i++ {
		if kline_items[i].Percent >= limit_up_rate {
			result++
		}
	}

	return result
}

// @func 计算周期内出现在龙虎榜净买的次数
func CalculatePeriodStockLhbTimes(stock_item data_center.StockItem, kline_items []data_center.KlineItem, start int, end int) int64 {
	var result int64 = 0
	for i := start; i < end; i++ {
		lhb_items := data_center.GetLhbStockItems(kline_items[i].Date)

		for _, lhb_item := range lhb_items {
			if lhb_item.Stock.Symbol != stock_item.Symbol {
				continue
			}
			if lhb_item.NetBuyTotal > 0 {
				result++
			}
		}
	}

	return result
}
