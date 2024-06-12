package data_center

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type KlineType = int64

const (
	KlineType_Day     KlineType = 0 // 日k
	KlineType_Week    KlineType = 1 // 周k
	KlineType_Month   KlineType = 2 // 月k
	KlineType_Quarter KlineType = 3 // 季k
	KlineType_Year    KlineType = 4 // 年k
	KlineType_1Min    KlineType = 5 // 1min 下面类似
	KlineType_5Min    KlineType = 6
	KlineType_15Min   KlineType = 7
	KlineType_30Min   KlineType = 8
	KlineType_60Min   KlineType = 9
	KlineType_120Min  KlineType = 10
)

type KlineItem struct {
	Timestamp    int64   // 时间戳，单位s
	Date         string  // 字符串日期 20240523
	Volume       uint64  // 成交量 单位股
	Open         int64   // 开盘价 单位分
	High         int64   // 最高价 单位分
	Low          int64   // 最低价 单位分
	Close        int64   // 收盘价 单位分
	TurnoverRate float64 // 换手率
	Amount       uint64  // 成交额 单位分
	Chg          int64   // 涨跌额 单位分
	Percent      float64 // 涨跌幅 已经乘了100
	EntityHigh   int64   // 实体最高价
	EntityLow    int64   // 实体最低价
	RSI6         float64 // RSI6天指标
}

type StockItem struct {
	Name   string // 股票名
	Symbol string // 股票代号 sz399001 sh000001
}

type LhbStockItem struct {
	Name        string         // 股票名
	Symbol      string         // 股票代号 sz399001 sh000001
	Date        string         // 上榜日期
	BuyTotal    int64          // 买入 单位分
	SellTotal   int64          // 卖出 单位分
	NetBuyTotal int64          // 净买入 单位分
	Percent     float64        // 涨跌幅 已经乘了100
	LhbBranch   LhbStockBranch // 龙虎榜机构数据
}

type LhbStockBranch struct {
	BuyBranches  []LhbStockBranchItem
	SellBranches []LhbStockBranchItem
}

type LhbStockBranchItem struct {
	BranchCode  string // 机构代码
	Branchname  string // 机构名
	BuyTotal    int64  // 买入 单位分
	SellTotal   int64  // 卖出 单位分
	NetBuyTotal int64  // 净买入 单位分
}

type JRJQuotKline struct {
	Data    JRJData `json:"data"`
	Msg     string  `json:"msg"`
	Retcode int64   `json:"retcode"`
}

type JRJData struct {
	Kline       []JRJKline `json:"kline"`
	Count       uint64     `json:"count"`
	NSecurityID uint64     `json:"nSecurityID"`
}

type JRJKline struct {
	LlValue     uint64 `json:"llValue"`     // 成交额
	LlVolume    uint64 `json:"llVolume"`    // 成交量
	NHighPx     int64  `json:"nHighPx"`     // 最高价
	NLastPx     int64  `json:"nLastPx"`     // 收盘价
	NLowPx      int64  `json:"nLowPx"`      // 最低价
	NOpenPx     int64  `json:"nOpenPx"`     // 开盘价
	NTime       uint64 `json:"nTime"`       // 时间 20240523
	NPreClosePx int64  `json:"nPreClosePx"` // 前一天收盘价
}

// @func 读写股票k线数据缓存
// @index 缓存文件索引
// @cache 待缓存的数据或者读取的缓存
// @read true=读 false=写
func GetKlineItemsGetOrSet(key string, cache *string, read bool) {
	work_dir, _ := os.Getwd()

	cache_path := fmt.Sprintf("%s\\%s\\%s.cache", work_dir, "data_center\\cache\\GetKlineItems", key)

	if read {
		file, err := os.ReadFile(cache_path)
		if err != nil {
			return
		}
		*cache = string(file)
	} else {
		cache_byte := []byte(*cache)
		os.WriteFile(cache_path, cache_byte, 0)
	}
}

// @func 获取最新的k线元素
// @kline_type k线类型
// @symbol 股票代号 sz399001 sh000001
// @count 数据总量
// @return 按照时间顺序返回每个k线图
var kline_items_map = make(map[string][]KlineItem)

func GetKlineItems(kline_type KlineType, symbol string, count uint64) []KlineItem {
	key := fmt.Sprintf("%s_%s_%d_%d", time.Now().Local().Format("20060102"), symbol, kline_type, count)
	kline_items, ok := kline_items_map[key]
	if ok {
		return kline_items
	}

	kTypeStringMap := map[KlineType]string{
		KlineType_Day:     "day",
		KlineType_Week:    "week",
		KlineType_Month:   "month",
		KlineType_Quarter: "quarter",
		KlineType_Year:    "year",
		KlineType_1Min:    "1m",
		KlineType_5Min:    "5m",
		KlineType_15Min:   "15m",
		KlineType_30Min:   "30m",
		KlineType_60Min:   "60m",
		KlineType_120Min:  "120m",
	}

	result := make([]KlineItem, 0)

	var body_string string
	GetKlineItemsGetOrSet(key, &body_string, true)
	if len(body_string) <= 0 {
		real_symbol := ""
		if strings.HasPrefix(strings.ToLower(symbol), "sh") {
			real_symbol = fmt.Sprintf("%d%s", 1, symbol[2:])
		} else if strings.HasPrefix(strings.ToLower(symbol), "sz") {
			real_symbol = fmt.Sprintf("%d%s", 2, symbol[2:])
		}

		format_string := "https://gateway.jrj.com/quot-kline?format=json&securityId=%s&type=%s&direction=left&range.num=%d"

		url := fmt.Sprintf(format_string, real_symbol, kTypeStringMap[kline_type], count)
		req, _ := http.NewRequest(http.MethodGet, url, nil)

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			return result
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		body_string = string(body)
		GetKlineItemsGetOrSet(key, &body_string, false)
	}

	body_byte := []byte(body_string)

	var quot_kline JRJQuotKline
	json.Unmarshal(body_byte, &quot_kline)

	for index, iter := range quot_kline.Data.Kline {
		if index == 0 && iter.NPreClosePx == 0 {
			iter.NPreClosePx = iter.NLastPx
		}

		if iter.NOpenPx == 0 || iter.NLastPx == 0 || iter.NHighPx == 0 || iter.NLowPx == 0 || iter.NTime <= 0 || iter.NPreClosePx <= 0 {
			break
		}

		var item KlineItem
		item.Date = fmt.Sprintf("%d", iter.NTime)
		date, _ := time.ParseInLocation("20060102", item.Date, time.Local)
		item.Timestamp = date.Unix()

		if len(item.Date) < 8 {
			break
		}

		item.Volume = iter.LlVolume
		item.Open = iter.NOpenPx / 100
		item.High = iter.NHighPx / 100
		item.Low = iter.NLowPx / 100
		item.Close = iter.NLastPx / 100
		item.TurnoverRate = 0.0
		item.Amount = iter.LlValue / 100
		item.Chg = (item.Close - item.Open)
		item.Percent = float64(item.Close-(iter.NPreClosePx/100)) / float64(iter.NPreClosePx/100) * 100

		item.EntityHigh = item.Open
		item.EntityLow = item.Close
		if item.Open < item.Close {
			item.EntityHigh = item.Close
			item.EntityLow = item.Open
		}
		result = append(result, item)
	}

	RSI(result)

	kline_items_map[key] = result

	return result
}

func RSI(kline_items []KlineItem) {
	rsi6_red, rsi6_all := 1e-6, 1e-6
	for index := 0; index < len(kline_items); index++ {
		gap := 0.0
		if index > 0 {
			gap = float64(kline_items[index].Close - kline_items[index-1].Close)
		}

		rsi6_red = (float64(6-1)*rsi6_red + math.Max(gap, 0.0)) / float64(6)
		rsi6_all = (float64(6-1)*rsi6_all + math.Abs(gap)) / float64(6)
		kline_items[index].RSI6 = rsi6_red / rsi6_all * 100
	}
}

type JRJHqs struct {
	Data JRJHqsData `json:"data"`
	Msg  string     `json:"msg"`
	Code int64      `json:"code"`
}

type JRJHqsData struct {
	Hqs []JRJHqsDataItem `json:"hqs"`
}

type JRJHqsDataItem struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Mkt  uint64 `json:"mkt"`
}

// @func 读写全部股票数据缓存
// @index 缓存文件索引
// @cache 待缓存的数据或者读取的缓存
// @read true=读 false=写
func GetWholeStockItemsCacheGetOrSet(key string, cache *string, read bool) {
	work_dir, _ := os.Getwd()
	cache_path := fmt.Sprintf("%s\\%s\\%s.cache", work_dir, "data_center\\cache\\GetWholeStockItems", key)

	if read {
		file, err := os.ReadFile(cache_path)
		if err != nil {
			return
		}
		*cache = string(file)
	} else {
		cache_byte := []byte(*cache)
		os.WriteFile(cache_path, cache_byte, 0)
	}
}

var stock_items_map = make(map[string][]StockItem)

func GetStockItems(cat uint64) []StockItem {
	key := fmt.Sprintf("%s_%d", time.Now().Local().Format("20060102"), cat)

	stock_items, ok := stock_items_map[key]
	if ok {
		return stock_items
	}
	url := "https://gateway.jrj.com/quot-feed/category_hqs"

	result := make([]StockItem, 0)

	for i := 0; i < 300; i++ {
		start := uint64(i * 20)

		file_key := fmt.Sprintf("%s_%d", key, i)

		var body_string string
		GetWholeStockItemsCacheGetOrSet(file_key, &body_string, true)
		if len(body_string) <= 0 {
			format_string := "{\"start\":%d,\"num\":20,\"currentPage\":%d,\"env\":[1,2],\"cat\":%d,\"column\":5,\"sort\":2}"
			json_payload := fmt.Sprintf(format_string, start, i+1, cat)
			payload := strings.NewReader(json_payload)
			req, _ := http.NewRequest(http.MethodPost, url, payload)
			req.Header.Add("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			body_string = string(body)

			GetWholeStockItemsCacheGetOrSet(file_key, &body_string, false)
		}

		body_byte := []byte(body_string)

		var hqs JRJHqs
		json.Unmarshal(body_byte, &hqs)

		for _, iter := range hqs.Data.Hqs {
			var item StockItem
			item.Name = iter.Name
			if iter.Mkt == 1 {
				item.Symbol = fmt.Sprintf("SH%s", iter.Code)
			} else if iter.Mkt == 2 {
				item.Symbol = fmt.Sprintf("SZ%s", iter.Code)
			} else {
				continue
			}
			result = append(result, item)
		}
	}

	stock_items_map[key] = result

	fmt.Printf("GetWholeStockItems, Total %d\n", len(result))

	return result
}

// @func 获取全A股
// @return StockItem 返回的全部股票
func GetWholeStockItems() []StockItem {
	return GetStockItems(1)
}

// @func 获取全A Etf
// @return StockItem 返回的全部etf
func GetWholeEtfStockItems() []StockItem {
	return GetStockItems(6)
}

type JRJLhb struct {
	Code int64      `json:"code"`
	Msg  string     `json:"msg"`
	Data JRJLhbData `json:"data"`
}

type JRJLhbData struct {
	Total     uint64         `json:"total"`
	LhbStocks []JRJLhbStocks `json:"lhbStocks"`
}

type JRJLhbStocks struct {
	Market         uint64       `json:"market"`
	Chg            float64      `json:"chg"`
	StockCode      string       `json:"stockCode"`
	StockName      string       `json:"stockName"`
	BuyValueTotal  float64      `json:"buyValueTotal"`
	SellValueTotal float64      `json:"sellValueTotal"`
	NetValueTotal  float64      `json:"netValueTotal"`
	LhbBranch      JRJLhbBranch `json:"lhbBranch"`
}

type JRJLhbBranch struct {
	BuyBranches  []JRJLhbBranchItem `json:"buyBranches"`
	SellBranches []JRJLhbBranchItem `json:"sellBranches"`
}

type JRJLhbBranchItem struct {
	BranchCode string  `json:"branchCode"`
	BranchName string  `json:"branchName"`
	BuyValue   float64 `json:"buyValue"`
	SellValue  float64 `json:"sellValue"`
	NetValue   float64 `json:"netValue"`
}

// @func 读写龙虎榜数据缓存
// @index 缓存文件索引
// @cache 待缓存的数据或者读取的缓存
// @read true=读 false=写
func GetLhbStockItemsGetOrSet(key string, cache *string, read bool) {
	work_dir, _ := os.Getwd()
	cache_path := fmt.Sprintf("%s\\%s\\%s.cache", work_dir, "data_center\\cache\\GetLhbStockItems", key)

	if read {
		file, err := os.ReadFile(cache_path)
		if err != nil {
			return
		}
		*cache = string(file)
	} else {
		cache_byte := []byte(*cache)
		os.WriteFile(cache_path, cache_byte, 0)
	}
}

// @func 获取某一天的龙虎榜 20240531
// @return LhbStockItem 返回当天的龙虎榜数据
var lhb_stock_items_map = make(map[string][]LhbStockItem)

func GetLhbStockItems(date string) []LhbStockItem {
	key := date
	lhb_stock_items, ok := lhb_stock_items_map[key]
	if ok {
		return lhb_stock_items
	}

	url := "https://gateway.jrj.com/quot-dc/lhb/stocklist"

	result := make([]LhbStockItem, 0)

	unix, _ := time.Parse("20060102", date)
	format_date := unix.Local().Format("2006-01-02")

	var body_string string
	GetLhbStockItemsGetOrSet(key, &body_string, true)
	if len(body_string) <= 0 {
		format_string := "{\"queryFlag\":2,\"endDate\":\"%s\",\"pageNum\":0,\"pageSize\":0}"
		json_payload := fmt.Sprintf(format_string, format_date)

		payload := strings.NewReader(json_payload)
		req, _ := http.NewRequest(http.MethodPost, url, payload)
		req.Header.Add("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			return result
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		body_string = string(body)

		GetLhbStockItemsGetOrSet(key, &body_string, false)
	}

	body_byte := []byte(body_string)

	var lhb JRJLhb
	json.Unmarshal(body_byte, &lhb)

	for _, iter := range lhb.Data.LhbStocks {
		var item LhbStockItem
		item.Name = iter.StockName
		if iter.Market == 1 {
			item.Symbol = fmt.Sprintf("SH%s", iter.StockCode)
		} else if iter.Market == 2 {
			item.Symbol = fmt.Sprintf("SZ%s", iter.StockCode)
		} else {
			continue
		}

		item.Date = date
		item.BuyTotal = int64(iter.BuyValueTotal * 10000 * 100)
		item.SellTotal = int64(iter.SellValueTotal * 10000 * 100)
		item.NetBuyTotal = int64(iter.NetValueTotal * 10000 * 100)
		item.Percent = iter.Chg * 100.0

		if item.BuyTotal <= 0 && item.SellTotal <= 0 && item.NetBuyTotal <= 0 {
			continue
		}

		for _, buy_iter := range iter.LhbBranch.BuyBranches {
			var buy_item LhbStockBranchItem

			buy_item.BranchCode = buy_iter.BranchCode
			buy_item.Branchname = buy_iter.BranchName
			buy_item.BuyTotal = int64(buy_iter.BuyValue * 10000 * 100)
			buy_item.SellTotal = int64(buy_iter.SellValue * 10000 * 100)
			buy_item.NetBuyTotal = int64(buy_iter.NetValue * 10000 * 100)

			if buy_item.BuyTotal <= 0 && buy_item.SellTotal <= 0 && buy_item.NetBuyTotal <= 0 {
				continue
			}

			item.LhbBranch.BuyBranches = append(item.LhbBranch.BuyBranches, buy_item)
		}

		for _, sell_iter := range iter.LhbBranch.SellBranches {
			var sell_item LhbStockBranchItem

			sell_item.BranchCode = sell_iter.BranchCode
			sell_item.Branchname = sell_iter.BranchName
			sell_item.BuyTotal = int64(sell_iter.BuyValue * 10000 * 100)
			sell_item.SellTotal = int64(sell_iter.SellValue * 10000 * 100)
			sell_item.NetBuyTotal = int64(sell_iter.NetValue * 10000 * 100)

			if sell_item.BuyTotal <= 0 && sell_item.SellTotal <= 0 && sell_item.NetBuyTotal <= 0 {
				continue
			}

			item.LhbBranch.SellBranches = append(item.LhbBranch.SellBranches, sell_item)
		}

		result = append(result, item)
	}

	lhb_stock_items_map[key] = result

	return result
}
