package main

import (
	"fmt"
	"os"

	in "shoper_updater/input"
	sh "shoper_updater/shoper"
)

func main() {
	os.MkdirAll("log/", 0700)

	fmt.Println("=== Stan_mag.txt ===")

	stanMag, errCount, err := in.ParseStanMag("data/Stan_mag.txt")
	if err != nil {
		fmt.Println("can't parse ./data/Stan_mag.txt, aborting...")
		fmt.Println(err)
		return
	}
	fmt.Printf("products\t%6d\n", len(stanMag))
	fmt.Printf("errors\t\t%6d\n", errCount)
	errRate := errCount * 100 / len(stanMag)
	fmt.Printf("error rate\t%5d%%\n", errRate)

	if errRate > 30 {
		fmt.Println("error rate above 30%, aborting...")
		return
	}

	url := os.Getenv("SHOPER_URL")
	login := os.Getenv("SHOPER_LOGIN")
	pass := os.Getenv("SHOPER_PASS")

	s, err := sh.NewSession(url, login, pass)
	if err != nil {
		fmt.Println("error while creating session, aborting...")
		fmt.Println(err)
		return
	}

	fmt.Printf("\n=== %s ===\n", s.Domain)

	stockList, err := s.GetStockList()
	if err != nil {
		fmt.Println("error while getting stock page, aborting...")
		fmt.Println(err)
		return
	}

	stocks, err := sh.GetStockMap(stockList)
	if err != nil {
		fmt.Println("error while preparing stock map, aborting...")
		fmt.Println(err)
		return
	}
	fmt.Printf("products\t%6d\n", len(stocks))

	stocks = sh.GetStanMagStock(stocks, stanMag)
	fmt.Printf("in Stan_mag.txt\t%6d\n", len(stocks))

	stocks, err = sh.GetUpdateStock(stocks)
	if err != nil {
		fmt.Println("error while preparing update map, aborting...")
		fmt.Println(err)
		return
	}
	fmt.Printf("to update\t%6d\n", len(stocks))

	s.LogFd.Close()
}
