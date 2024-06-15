package technical_analysis

import (
	"fmt"
	"math"

	"github.com/hsuloong/stock_speculation/data_center"
)

// @func 相对低点分析
func StartEtfLofLowAnalysis() {
	whole_etf_lof_stock_items := data_center.GetWholeEtfStockItems()
	// whole_etf_lof_stock_items := data_center.GetWholeLofStockItems()
	// whole_etf_lof_stock_items := data_center.GetWholeIndexStockItems()
	// whole_etf_lof_stock_items := data_center.GetWholeIndustryStockItems()
	// whole_etf_lof_stock_items := data_center.GetWholeConceptStockItems()

	const kMaxRate float64 = 10.0
	const kYearTradeDays int = 250
	const kMaxYears int = 14
	const kTargetYear int = 5

	for _, iter := range whole_etf_lof_stock_items {
		kline_items := data_center.GetKlineItems(data_center.KlineType_Day, iter, 250*14)
		kline_items_len := len(kline_items)

		target_lowest_rate := math.MaxFloat64
		target_relatively_rate := math.MaxFloat64
		result := ""
		for i := 1; i <= kMaxYears; i++ {
			days_before := i * kYearTradeDays
			if kline_items_len < days_before {
				continue
			}

			loop_lowest_rate := CalculatePeriodToLowestPercent(kline_items, kline_items_len-days_before, kline_items_len, kline_items_len-1)
			loop_relatively_rate := CalculatePeriodRelativelyPercent(kline_items, kline_items_len-days_before, kline_items_len, kline_items_len-1)

			if kTargetYear == i {
				target_lowest_rate = loop_lowest_rate
				target_relatively_rate = loop_relatively_rate
			}

			result += fmt.Sprintf("最近%d年 %f|%f ", i, loop_lowest_rate, loop_relatively_rate)
		}

		if target_lowest_rate <= kMaxRate && target_relatively_rate <= kMaxRate && kline_items[kline_items_len-1].Amount >= 1e7*10000 {
			fmt.Printf("%s(%s) %s\n", iter.Name, iter.Symbol, result)
		}
	}
}

// @func 波动率分析
func StartEtfLofVolatilityAnalysis() {
	const kMinRate float64 = 2.0
	const kMonthTradeDays int = 21
	const kMaxMonths int = 6
	const kTargetMonth int = 3
	const kMinTradeAmount uint64 = 1e8 * 100 / 2
	const kMinAvgTradeDays int = 5

	whole_etf_lof_stock_items := data_center.GetWholeEtfStockItems()
	for _, iter := range whole_etf_lof_stock_items {
		kline_items := data_center.GetKlineItems(data_center.KlineType_Day, iter, 250*14)
		kline_items_len := len(kline_items)

		target_volatility := 0.0
		result := ""
		for i := 1; i <= kMaxMonths; i++ {
			days_before := i * kMonthTradeDays
			if kline_items_len < days_before {
				continue
			}

			loop_volatility := CalculatePeriodVolatility(kline_items, kline_items_len-days_before, kline_items_len)

			if kTargetMonth == i {
				target_volatility = loop_volatility
			}

			result += fmt.Sprintf("最近%d月 %f ", i, loop_volatility)
		}

		var period_amount_avg uint64 = 0
		if kline_items_len >= kMinAvgTradeDays {
			period_amount_avg = uint64(CalculatePeriodAvgAmount(kline_items, kline_items_len-kMinAvgTradeDays, kline_items_len))
		} else {
			period_amount_avg = uint64(CalculatePeriodAvgAmount(kline_items, 0, kline_items_len))
		}

		if target_volatility >= kMinRate && period_amount_avg >= kMinTradeAmount {
			fmt.Printf("%s(%s) %s\n", iter.Name, iter.Symbol, result)
		}
	}
}

// @func 网格交易分析
func StartEtfLofGridTradingAnalysis() {
	const kMonthTradeDays int = 21
	const kMaxMonths int = 6
	const kTargetMonth int = 3
	const kMinRate float64 = -100.0
	const kMinTradeAmount uint64 = 1e5 * 10000
	const kMinAvgTradeDays int = 5

	whole_etf_lof_stock_items := data_center.GetWholeEtfStockItems()
	for _, iter := range whole_etf_lof_stock_items {
		kline_items := data_center.GetKlineItems(data_center.KlineType_Day, iter, 250*14)
		kline_items_len := len(kline_items)

		target_profit_rate := -100.0
		result := ""
		for i := 1; i <= kMaxMonths; i++ {
			days_before := i * kMonthTradeDays
			if kline_items_len < days_before {
				continue
			}

			loop_profit_rate := CalculateGridTradingProfit(kline_items, kline_items_len-days_before, kline_items_len, 0.01)

			if kTargetMonth == i {
				target_profit_rate = loop_profit_rate
			}

			result += fmt.Sprintf("最近%d月 %f ", i, loop_profit_rate)
		}

		var period_amount_avg uint64 = 0
		if kline_items_len >= kMinAvgTradeDays {
			period_amount_avg = uint64(CalculatePeriodAvgAmount(kline_items, kline_items_len-kMinAvgTradeDays, kline_items_len))
		} else {
			period_amount_avg = uint64(CalculatePeriodAvgAmount(kline_items, 0, kline_items_len))
		}

		if target_profit_rate >= kMinRate && period_amount_avg >= kMinTradeAmount {
			fmt.Printf("%s(%s) %s\n", iter.Name, iter.Symbol, result)
		}
	}
}
