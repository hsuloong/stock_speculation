package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	technical_analysis "github.com/hsuloong/stock_speculation/technical_analysis"
)

var func_name = flag.String("f", "StartBacktesting", "运行的函数")

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("%s -f 函数名称\n", os.Args[0])
		return
	}
	flag.Parse()

	start := time.Now().Local().Unix()
	if *func_name == "StartBacktesting" {
		technical_analysis.StartBacktesting()
	} else if *func_name == "StartSelectStock" {
		technical_analysis.StartSelectStock()
	} else if *func_name == "StartSelectEtf" {
		technical_analysis.StartSelectEtf()
	} else if *func_name == "StartLhbAnalysis" {
		technical_analysis.StartLhbAnalysis()
	} else if *func_name == "StartEtfLofLowAnalysis" {
		technical_analysis.StartEtfLofLowAnalysis()
	} else if *func_name == "StartEtfLofVolatilityAnalysis" {
		technical_analysis.StartEtfLofVolatilityAnalysis()
	} else if *func_name == "StartEtfLofGridTradingAnalysis" {
		technical_analysis.StartEtfLofGridTradingAnalysis()
	} else if *func_name == "StartHotIndustryAnalysis" {
		technical_analysis.StartHotIndustryAnalysis()
	} else if *func_name == "StartHotIndustryHotStockAnalysis" {
		technical_analysis.StartHotIndustryHotStockAnalysis()
	} else if *func_name == "StartHotIndustryLhbAnalysis" {
		technical_analysis.StartHotIndustryLhbAnalysis()
	} else if *func_name == "StartHotIndustryLhbHotStockAnalysis" {
		technical_analysis.StartHotIndustryLhbHotStockAnalysis()
	}
	end := time.Now().Local().Unix()
	fmt.Printf("Total Cost: %ds\n", end-start)
}
