package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	technic_alanalysis "github.com/hsuloong/stock_speculation/technical_analysis"
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
		technic_alanalysis.StartBacktesting()
	} else if *func_name == "StartSelectStock" {
		technic_alanalysis.StartSelectStock()
	}
	end := time.Now().Local().Unix()
	fmt.Printf("Total Cost: %ds\n", end-start)
}
