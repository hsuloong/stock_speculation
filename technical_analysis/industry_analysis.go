package technical_analysis

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hsuloong/stock_speculation/data_center"
)

func StartHotIndustryAnalysis() {
	const kYearTradeDays int = 250
	const kMaxYears int = 14
	const kTargetYear int = 5
	const kMinAvgLimitUpTimes float64 = 3.0

	whole_industries := data_center.GetWholeIndustryStockItems()
	for _, industry_item := range whole_industries {
		industry_stock_items := data_center.GetIndexContainStockItems(industry_item)
		industry_stock_items_len := len(industry_stock_items)

		var target_limit_up_times int64 = 0
		result := ""
		for i := 1; i <= kMaxYears; i++ {
			var loop_limit_up_times int64 = 0
			for _, stock_item := range industry_stock_items {
				kline_items := data_center.GetKlineItems(data_center.KlineType_Day, stock_item, 250*14)
				kline_items_len := len(kline_items)

				days_before := i * kYearTradeDays
				if kline_items_len < days_before {
					continue
				}

				loop_limit_up_times += CalculateStockLimitUpTimes(kline_items, kline_items_len-days_before, kline_items_len, 9.0)
			}

			if kTargetYear == i {
				target_limit_up_times = loop_limit_up_times
			}

			result += fmt.Sprintf("最近%d年 %f ", i, float64(loop_limit_up_times)/float64(industry_stock_items_len*i))
		}

		if float64(target_limit_up_times)/float64(industry_stock_items_len*kTargetYear) >= kMinAvgLimitUpTimes {
			fmt.Printf("%s(%s) %s\n", industry_item.Name, industry_item.Symbol, result)
		}
	}
}

func StartHotIndustryHotStockAnalysis() {
	const kMaxDaysBefore int = 250 * 1
	const kMaxStocks int = 5

	hot_idustries := "BK20720500, BK20720600, BK20461100, BK20430300, BK20720700, BK20280600, BK20280500"

	whole_industries := data_center.GetWholeIndustryStockItems()
	for _, industry_item := range whole_industries {
		if !strings.Contains(hot_idustries, industry_item.Symbol) {
			continue
		}
		industry_stock_items := data_center.GetIndexContainStockItems(industry_item)

		limit_up_stock := make([]string, 0)
		for _, stock_item := range industry_stock_items {
			kline_items := data_center.GetKlineItems(data_center.KlineType_Day, stock_item, 250*14)
			kline_items_len := len(kline_items)

			if kline_items_len < kMaxDaysBefore {
				continue
			}

			loop_limit_up_times := CalculateStockLimitUpTimes(kline_items, kline_items_len-kMaxDaysBefore, kline_items_len, 9.0)
			if loop_limit_up_times > 0 {
				limit_up_stock = append(limit_up_stock, fmt.Sprintf("%d_%s(%s)", loop_limit_up_times, stock_item.Name, stock_item.Symbol))
			}
		}

		sort.Slice(limit_up_stock, func(i, j int) bool {
			i_value, _ := strconv.Atoi(limit_up_stock[i][0:strings.Index(limit_up_stock[i], "_")])
			j_value, _ := strconv.Atoi(limit_up_stock[j][0:strings.Index(limit_up_stock[j], "_")])
			return i_value > j_value
		})

		fmt.Printf("%s: \n", industry_item.Name)
		for i := 0; i < kMaxStocks && i < len(limit_up_stock); i++ {
			fmt.Printf("%s\n", limit_up_stock[i])
		}
	}
}

func StartHotIndustryLhbAnalysis() {
	const kYearTradeDays int = 250
	const kMaxYears int = 14
	const kTargetYear int = 5
	const kMinAvgLhbTimes float64 = 0.0

	whole_industries := data_center.GetWholeIndustryStockItems()
	for _, industry_item := range whole_industries {
		industry_stock_items := data_center.GetIndexContainStockItems(industry_item)
		industry_stock_items_len := len(industry_stock_items)

		var target_lhb_times int64 = 0
		result := ""
		for i := 1; i <= kMaxYears; i++ {
			var loop_lhb_times int64 = 0
			for _, stock_item := range industry_stock_items {
				kline_items := data_center.GetKlineItems(data_center.KlineType_Day, stock_item, 250*14)
				kline_items_len := len(kline_items)

				days_before := i * kYearTradeDays
				if kline_items_len < days_before {
					continue
				}

				loop_lhb_times += CalculatePeriodStockLhbTimes(stock_item, kline_items, kline_items_len-days_before, kline_items_len)
			}

			if kTargetYear == i {
				target_lhb_times = loop_lhb_times
			}

			result += fmt.Sprintf("最近%d年 %f ", i, float64(loop_lhb_times)/float64(industry_stock_items_len*i))
		}

		if float64(target_lhb_times)/float64(industry_stock_items_len*kTargetYear) >= kMinAvgLhbTimes {
			fmt.Printf("%s(%s) %s\n", industry_item.Name, industry_item.Symbol, result)
		}
	}
}

func StartHotIndustryLhbHotStockAnalysis() {
	const kMaxDaysBefore int = 250 * 1
	const kMaxStocks int = 5

	hot_idustries := "BK20720500, BK20720600, BK20461100, BK20430300, BK20720700, BK20280600, BK20280500, BK20110200, BK20510100, BK20280300, BK20240400, BK20720400"

	whole_industries := data_center.GetWholeIndustryStockItems()
	for _, industry_item := range whole_industries {
		if !strings.Contains(hot_idustries, industry_item.Symbol) {
			continue
		}
		industry_stock_items := data_center.GetIndexContainStockItems(industry_item)

		lhb_stock := make([]string, 0)
		for _, stock_item := range industry_stock_items {
			kline_items := data_center.GetKlineItems(data_center.KlineType_Day, stock_item, 250*14)
			kline_items_len := len(kline_items)
			if kline_items_len < kMaxDaysBefore {
				continue
			}

			loop_lhb_times := CalculatePeriodStockLhbTimes(stock_item, kline_items, kline_items_len-kMaxDaysBefore, kline_items_len)
			if loop_lhb_times > 0 {
				lhb_stock = append(lhb_stock, fmt.Sprintf("%d_%s(%s)", loop_lhb_times, stock_item.Name, stock_item.Symbol))
			}
		}

		sort.Slice(lhb_stock, func(i, j int) bool {
			i_value, _ := strconv.Atoi(lhb_stock[i][0:strings.Index(lhb_stock[i], "_")])
			j_value, _ := strconv.Atoi(lhb_stock[j][0:strings.Index(lhb_stock[j], "_")])
			return i_value > j_value
		})

		fmt.Printf("%s: \n", industry_item.Name)
		for i := 0; i < kMaxStocks && i < len(lhb_stock); i++ {
			fmt.Printf("%s\n", lhb_stock[i])
		}
	}
}
