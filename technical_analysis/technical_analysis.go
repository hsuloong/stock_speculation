package technic_alanalysis

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/hsuloong/stock_speculation/data_center"
)

func StartBacktesting() {
	fmt.Println("StartBacktesting")

	const kMaxSellDays int = 10              // 卖出距离买入日
	const kMaxMoneyPerTrade float64 = 100000 // 单次交易金额
	const kTax float64 = 0.2 / 100.0         // 交易费率 简单处理下

	for year, year_gap := 2010, 1; year <= time.Now().Local().Year(); year += year_gap {
		year_begin := year
		year_end := year + year_gap
		fmt.Printf("\nFrom Year %d To %d:\n", year, year+year_gap)

		win := make([]uint64, kMaxSellDays)
		loss := make([]uint64, kMaxSellDays)
		account_money := make([][20351231]float64, kMaxSellDays)

		whole_stock_items := data_center.GetWholeStockItems()
		for _, iter := range whole_stock_items {
			kline_items := data_center.GetKlineItems(data_center.KlineType_Day, iter.Symbol, 250*14)
			kline_items_len := len(kline_items)

			for i := 0; i < kline_items_len; i++ {
				kline_item_date, _ := strconv.Atoi(kline_items[i].Date)
				kline_item_year := kline_item_date / 10000

				// 按照年来回测
				if kline_item_year < year_begin || kline_item_year >= year_end {
					continue
				}

				// 判断命中条件
				is_match := IsIsVenusPattern(kline_items, i)
				if !is_match {
					continue
				}

				// 最后一个item无法计算买卖点
				if i+1 >= kline_items_len {
					continue
				}

				buy_index := i + 1 // 买入点 | 开盘买入
				buy_date, _ := strconv.Atoi(kline_items[buy_index].Date)

				// 遍历多个卖点
				for j := 0; j < kMaxSellDays; j++ {
					sell_index := i + j + 2 // 卖出点 | 收盘卖出
					if sell_index >= kline_items_len {
						break
					}
					sell_date, _ := strconv.Atoi(kline_items[sell_index].Date)
					rate := float64(kline_items[sell_index].Close) / float64(kline_items[buy_index].Open)

					// 记录胜率
					if rate > 1.0 {
						win[j]++
					} else {
						loss[j]++
					}

					// 记录交易金额
					account_money[j][buy_date] -= kMaxMoneyPerTrade * (1.0 + kTax)
					account_money[j][sell_date] += kMaxMoneyPerTrade * rate * (1.0 - kTax)
				}
			}
		}

		// 计算胜率
		win_rate := make([]float64, kMaxSellDays)
		for i := 0; i < kMaxSellDays; i++ {
			if win[i]+loss[i] > 0 {
				win_rate[i] = float64(win[i]) / float64(win[i]+loss[i]) * 100.0
			}
		}

		// 计算本金投入和收益
		invest_money := make([]float64, kMaxSellDays)
		profit := make([]float64, kMaxSellDays)
		profit_rate := make([]float64, kMaxSellDays)
		for i := 0; i < kMaxSellDays; i++ {
			invest_money[i] = (-MinSubArraySum(account_money[i][0:]))
			for j := 0; j < len(account_money[i]); j++ {
				profit[i] += account_money[i][j]
			}
			if invest_money[i] > 1.0e-3 {
				profit_rate[i] = profit[i] / invest_money[i] * 100.0
			}
		}

		// 打印结果
		fmt.Printf("胜次数: %v\n", win)
		fmt.Printf("败次数: %v\n", loss)
		fmt.Printf("胜率: %v\n", win_rate)
		fmt.Printf("本金投入: %v\n", invest_money)
		fmt.Printf("收益: %v\n", profit)
		fmt.Printf("收益率: %v\n", profit_rate)
	}
}

func MinSubArraySum(array []float64) float64 {
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

func StartSelectStock() {
	fmt.Println("StartSelectStock")
	whole_stock_items := data_center.GetWholeStockItems()

	for _, iter := range whole_stock_items {
		kline_items := data_center.GetKlineItems(data_center.KlineType_Day, iter.Symbol, 250*14)

		kline_items_len := len(kline_items)

		for i := kline_items_len - 1; i < kline_items_len; i++ {
			is_match := IsBullishEngulfingPattern(kline_items, i)
			if is_match {
				fmt.Printf("%s %d day %s\n", iter.Name, i, kline_items[i].Date)
			}
		}
	}
}

func IsDowntrend(kline_items []data_center.KlineItem, start int, end int) bool {
	if start < 0 || end > len(kline_items) {
		return false
	}

	for i := start + 1; i < end; i++ {
		if kline_items[i].Close > kline_items[i-1].Close {
			return false
		}
	}
	return true
}

// @func 判断是否是锤子线
// @kline_items k线数组
// @index 判定日期的数组索引
// @return 是否是锤子线
func IsHammerLinePattern(kline_items []data_center.KlineItem, index int) bool {
	kline_item := kline_items[index]

	l1 := kline_item.High - kline_item.EntityHigh
	l2 := kline_item.EntityHigh - kline_item.EntityLow
	l3 := kline_item.EntityLow - kline_item.Low

	// 上影线很短
	if float64(l1)*4.0 > float64(l2) {
		return false
	}

	// 下影线是实体的X倍
	if l3 <= 0 || l3 < l2*2 {
		return false
	}

	// 下降趋势
	if !IsDowntrend(kline_items, index-5, index) {
		return false
	}

	return true
}

// @func 是否是看涨吞没
// @kline_items k线数组
// @index 判定k线元素索引
// @return 是否是看涨吞没
func IsBullishEngulfingPattern(kline_items []data_center.KlineItem, index int) bool {
	if index < 1 {
		return false
	}

	first_kline_item := kline_items[index-1]
	second_kline_item := kline_items[index]

	// 第一根阴线
	if first_kline_item.Close > first_kline_item.Open {
		return false
	}

	// 第二根阳线
	if second_kline_item.Close < second_kline_item.Open {
		return false
	}

	// 实体吞没
	if second_kline_item.EntityHigh <= first_kline_item.EntityHigh || second_kline_item.EntityLow >= first_kline_item.EntityLow {
		return false
	}

	// 第二天实体非常大
	if second_kline_item.Percent < 3.0 {
		return false
	}

	// 下降趋势
	if !IsDowntrend(kline_items, index-5, index) {
		return false
	}

	return true
}

// @func 是否是刺透形态
// @kline_items k线数组
// @index 判定k线元素索引
// @return 是否是刺透形态
func IsPiercingPattern(kline_items []data_center.KlineItem, index int) bool {
	if index < 1 {
		return false
	}

	first_kline_item := kline_items[index-1]
	second_kline_item := kline_items[index]

	// 第一根阴线
	if first_kline_item.Close > first_kline_item.Open {
		return false
	}

	// 第二根阳线
	if second_kline_item.Close < second_kline_item.Open {
		return false
	}

	// 刺透到50%以上
	if float64(second_kline_item.EntityHigh-first_kline_item.EntityLow)*2.0 < float64(first_kline_item.EntityHigh-first_kline_item.EntityLow) {
		return false
	}

	// 下降趋势
	if !IsDowntrend(kline_items, index-5, index) {
		return false
	}

	return true
}

// @func 是否是启明星形态
// @kline_items k线数组
// @index 判定k线元素索引
// @return 是否是启明星形态
func IsIsVenusPattern(kline_items []data_center.KlineItem, index int) bool {
	if index < 2 {
		return false
	}

	first_kline_item := kline_items[index-2]
	second_kline_item := kline_items[index-1]
	third_kline_item := kline_items[index]

	// 第一根阴线
	if first_kline_item.Close > first_kline_item.Open {
		return false
	}

	// 第一根长阴线 | 参数可调
	if float64(first_kline_item.EntityHigh-first_kline_item.EntityLow)/float64(first_kline_item.EntityLow) < 0.03 {
		return false
	}

	// 第二根跳空低开
	if first_kline_item.EntityLow < second_kline_item.EntityHigh {
		return false
	}

	// 第二根小实体 | 参数可调
	if float64(second_kline_item.EntityHigh-second_kline_item.EntityLow)/float64(second_kline_item.EntityLow) > 0.01 {
		return false
	}

	// 第三根跳空高开
	if third_kline_item.EntityLow < second_kline_item.EntityHigh {
		return false
	}

	// 第三根阳线
	if third_kline_item.Close < third_kline_item.Open {
		return false
	}

	// 第三根阳线推进到第一个内部 | 参数可调
	if float64(third_kline_item.EntityHigh-first_kline_item.EntityLow)/float64(first_kline_item.EntityHigh-first_kline_item.EntityLow) < 0.3 {
		return false
	}

	// 第三根长阳线 | 参数可调
	if float64(third_kline_item.EntityHigh-third_kline_item.EntityLow)/float64(third_kline_item.EntityLow) < 0.03 {
		return false
	}

	// 下降趋势
	if !IsDowntrend(kline_items, index-5, index) {
		return false
	}

	return true
}

// @func 是否是孕线形态
// @kline_items k线数组
// @index 判定k线元素索引
// @return 是否是孕线形态
func HaramiPattern(kline_items []data_center.KlineItem, index int) bool {
	if index < 1 {
		return false
	}

	first_kline_item := kline_items[index-1]
	second_kline_item := kline_items[index]

	// 第一根阴线
	if first_kline_item.Close > first_kline_item.Open {
		return false
	}

	// 第一天实体非常大
	if float64(first_kline_item.EntityHigh-first_kline_item.EntityLow)/float64(first_kline_item.EntityLow) < 0.03 {
		return false
	}

	// 第二根被第一根吞没
	if second_kline_item.EntityHigh >= first_kline_item.EntityHigh || second_kline_item.EntityLow <= first_kline_item.EntityLow {
		return false
	}

	// 下降趋势
	if !IsDowntrend(kline_items, index-5, index) {
		return false
	}

	return false
}

// @func 是否平头底
// @kline_items k线数组
// @index 判定k线元素索引
// @return 是否是否平头底形态
func FlatBottomPattern(kline_items []data_center.KlineItem, index int) bool {
	if index < 1 {
		return false
	}

	first_kline_item := kline_items[index-1]
	second_kline_item := kline_items[index]

	// 底部差不多
	if math.Abs(float64(first_kline_item.Low-second_kline_item.Low))/float64(first_kline_item.Low) > 0.005 {
		return false
	}

	// 下降趋势
	if !IsDowntrend(kline_items, index-5, index) {
		return false
	}

	return true
}
