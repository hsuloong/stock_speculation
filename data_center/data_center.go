package data_center

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type KlineType = int8

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
	Volume       uint64  // 成交量 单位手
	Open         int64   // 开盘价 单位分 下面一样
	High         int64   // 最高价
	Low          int64   // 最低价
	Close        int64   // 收盘价
	TurnoverRate float32 // 换手率
	Amount       uint64  // 成交额
	Chg          int64   // 涨跌额
	Percent      float32 // 涨跌幅 已经乘了100
	EntityHigh   int64   // 实体最高价
	EntityLow    int64   // 实体最低价
}

type StockItem struct {
	Name   string // 股票名
	Symbol string // 股票代号 sz399001 sh000001
}

type JRJQuotKline struct {
	Data    JRJData `json:"data"`
	Msg     string  `json:"msg"`
	Retcode int32   `json:"retcode"`
}

type JRJData struct {
	Kline       []JRJKline `json:"kline"`
	Count       uint32     `json:"count"`
	NSecurityID uint64     `json:"nSecurityID"`
}

type JRJKline struct {
	LlValue     uint64 `json:"llValue"`     // 成交额
	LlVolume    uint64 `json:"llVolume"`    // 成交量
	NHighPx     int64  `json:"nHighPx"`     // 最高价
	NLastPx     int64  `json:"nLastPx"`     // 收盘价
	NLowPx      int64  `json:"nLowPx"`      // 最低价
	NOpenPx     int64  `json:"nOpenPx"`     // 开盘价
	NTime       uint32 `json:"nTime"`       // 时间 20240523
	NPreClosePx int64  `json:"nPreClosePx"` // 前一天收盘价
}

// @func 读写股票k线数据缓存
// @index 缓存文件索引
// @cache 待缓存的数据或者读取的缓存
// @read true=读 false=写
func GetKlineItemsGetOrSet(symbol string, index uint32, cache *string, read bool) {
	work_dir, _ := os.Getwd()
	now := time.Now().Local().Format("20060102")

	cache_path := fmt.Sprintf("%s\\%s\\%s_%s_%d.cache", work_dir, "data_center\\cache\\GetKlineItems", symbol, now, index)

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
func GetKlineItems(kline_type KlineType, symbol string, count uint32) []KlineItem {
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
	GetKlineItemsGetOrSet(symbol, count, &body_string, true)
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
		GetKlineItemsGetOrSet(symbol, count, &body_string, false)
	}

	body_byte := []byte(body_string)

	var quot_kline JRJQuotKline
	json.Unmarshal(body_byte, &quot_kline)

	for _, iter := range quot_kline.Data.Kline {
		var item KlineItem
		item.Date = fmt.Sprintf("%d", iter.NTime)
		date, _ := time.ParseInLocation("20060102", item.Date, time.Local)
		item.Timestamp = date.Unix()

		item.Volume = iter.LlVolume
		item.Open = iter.NOpenPx / 100
		item.High = iter.NHighPx / 100
		item.Low = iter.NLowPx / 100
		item.Close = iter.NLastPx / 100
		item.TurnoverRate = 0.0
		item.Amount = iter.LlValue / 100
		item.Chg = (item.Close - item.Open)
		item.Percent = float32(item.Close-(iter.NPreClosePx/100)) / float32(iter.NPreClosePx/100) * 100

		item.EntityHigh = item.Open
		item.EntityLow = item.Close
		if item.Open < item.Close {
			item.EntityHigh = item.Close
			item.EntityLow = item.Open
		}

		if item.Open == 0 || item.Close == 0 || item.High == 0 || item.Low == 0 || len(item.Date) <= 0 {
			break
		}

		result = append(result, item)
	}

	return result
}

type JRJHqs struct {
	Data JRJHqsData `json:"data"`
	Msg  string     `json:"msg"`
	Code int32      `json:"code"`
}

type JRJHqsData struct {
	Hqs []JRJHqsDataItem `json:"hqs"`
}

type JRJHqsDataItem struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Mkt  uint8  `json:"mkt"`
}

// @func 读写全部股票数据缓存
// @index 缓存文件索引
// @cache 待缓存的数据或者读取的缓存
// @read true=读 false=写
func GetWholeStockItemsCacheGetOrSet(index uint32, cache *string, read bool) {
	work_dir, _ := os.Getwd()
	now := time.Now().Local().Format("20060102")
	cache_path := fmt.Sprintf("%s\\%s\\%s_%d.cache", work_dir, "data_center\\cache\\GetWholeStockItems", now, index)

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

// @func 获取全A股
// @return StockItem 返回的全部股票
func GetWholeStockItems() []StockItem {
	url := "https://gateway.jrj.com/quot-feed/category_hqs"

	result := make([]StockItem, 0)

	for i := 0; i < 300; i++ {
		start := uint32(i * 20)

		var body_string string
		GetWholeStockItemsCacheGetOrSet(start, &body_string, true)
		if len(body_string) <= 0 {

			format_string := "{\"start\":%d,\"num\":20,\"currentPage\":%d,\"env\":[1,2],\"cat\":1,\"column\":8,\"sort\":2}"
			json_payload := fmt.Sprintf(format_string, start, i+1)
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

			GetWholeStockItemsCacheGetOrSet(start, &body_string, false)
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

	fmt.Printf("GetWholeStockItems, Total %d\n", len(result))

	return result
}
