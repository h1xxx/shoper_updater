package main

import (
	"fmt"
	"os"

	in "shoper_updater/input"
	sh "shoper_updater/shoper"
)

func main() {
	os.MkdirAll("tmp/", 0700)

	stanMag, errCount, err := in.ParseStanMag("data/Stan_mag.txt")
	if err != nil {
		fmt.Println("can't open ./data/Stan_mag.txt, aborting...")
		return
	}
	fmt.Println("=== Stan_mag.txt ===")
	fmt.Printf("products:\t%d\n", len(stanMag))
	fmt.Printf("errors:\t\t%d\n", errCount)
	errRate := errCount * 100 / len(stanMag)
	fmt.Printf("error rate:\t%d%%\n", errRate)

	if errRate > 30 {
		fmt.Println("error rate above 30%, aborting...")
		return
	}

	url := os.Getenv("SHOPER_URL")
	login := os.Getenv("SHOPER_LOGIN")
	pass := os.Getenv("SHOPER_PASS")

	s, err := sh.NewSession(url, login, pass)
	if err != nil {
		fmt.Println("error while creating session, aboring...")
		fmt.Println(err)
		return
	}

	stockPage, err := s.GetStock()
	if err != nil {
		fmt.Println("error while getting stock page, aboring...")
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", len(stockPage.StockList))

	s.LogFd.Close()
}
