package main

import (
	"errors"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	in "shoper_updater/input"
	sh "shoper_updater/shoper"
)

func main() {
	os.MkdirAll("log/", 0700)

	// get credentials
	var shops []in.ShopT
	var shop in.ShopT
	var err error

	shop.Url = os.Getenv("SHOPER_URL")
	shop.Login = os.Getenv("SHOPER_LOGIN")
	shop.Pass = os.Getenv("SHOPER_PASS")

	if shop.Url != "" && shop.Login != "" && shop.Pass != "" {
		shops = append(shops, shop)
	} else {
		shops, err = in.ParseLoginConf()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// main loop
	lastTime := time.Now()
	mTime := lastTime
	msgWait := "retrying in 10 min."
	for {
		var errAny error

		fmt.Printf("\nstart time: %s\n",
			time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println("-------------------------------")

		// read config for product sets
		sets := getProductSets()

		// read current stock data
		stanMag, err := getStanMag()
		if err != nil {
			msg := "can't parse ./data/Stan_mag.txt"
			fmt.Printf("%s, %s\n%s\n", msg, msgWait, err)
			time.Sleep(10 * time.Minute)
			continue
		}

		// add stock info for product sets to stanMag data
		in.AddSets(stanMag, sets)

		// read data from each shop and update it
		for _, shop := range shops {
			err, errApi := shopUpdate(stanMag, shop)
			if err != nil {
				errAny = err
			}
			if errApi != nil {
				fmt.Println(errApi)
			}
		}

		// repeat the update if any errors where encountered
		if errAny != nil {
			msg := "error during an update"
			fmt.Printf("%s, %s\n%s\n", msg, msgWait, errAny)
			time.Sleep(10 * time.Minute)
			continue
		}

		// wait for new Stan_mag.txt file
		fmt.Println("\nwaiting for data/Stan_mag.txt update...")
		for mTime.Before(lastTime) || mTime == lastTime {
			fInfo, err := os.Stat("data/Stan_mag.txt")
			if err != nil {
				msg := "can't read data/Stan_mag.txt"
				fmt.Printf("%s, %s\n%s\n", msg, msgWait, err)
				time.Sleep(10 * time.Minute)
				continue
			}
			mTime = fInfo.ModTime()
			time.Sleep(1 * time.Second)
		}
		lastTime = mTime
	}
}

func getStanMag() (map[string]float64, error) {
	fmt.Println("\n=== Stan_mag.txt ===")

	stanMag, errCount, err := in.ParseStanMag("data/Stan_mag.txt")
	if err != nil {
		return stanMag, err
	}
	fmt.Printf("products\t%6d\n", len(stanMag))
	fmt.Printf("errors\t\t%6d\n", errCount)
	errRate := errCount * 100 / len(stanMag)
	fmt.Printf("error rate\t%5d%% (max 50%%)\n", errRate)

	if errRate > 50 {
		return stanMag, errors.New("error rate above 50%")
	}

	return stanMag, nil
}

func getProductSets() []in.ProductSetT {
	fmt.Println("\n=== product_set.conf ===")

	sets, errCount, err := in.ParseProductSets("etc/product_set.conf")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("sets\t\t%6d\n", len(sets))
	fmt.Printf("errors\t\t%6d\n", errCount)

	return sets
}

// first returned err is for general errors, second one for errors from
// api when updating stock info
func shopUpdate(stanMag map[string]float64, shop in.ShopT) (error, error) {
	fmt.Printf("\n=== %s ===\n", strings.Trim(fp.Base(shop.Url), "www."))

	s, err := sh.NewSession(shop.Url, shop.Login, shop.Pass)
	if err != nil {
		return err, nil
	}

	fmt.Println("getting info from shoper...")
	stockList, err := s.GetStockList()
	if err != nil {
		return err, nil
	}

	stocks := sh.GetStockMap(stockList)
	fmt.Printf("products\t%6d\n", len(stocks))

	stocks = sh.GetStanMagStock(stocks, stanMag)
	fmt.Printf("in Stan_mag.txt\t%6d\n", len(stocks))

	stocks, err = sh.GetUpdateStock(stocks)
	if err != nil {
		return err, nil
	}
	fmt.Printf("to update\t%6d\n", len(stocks))

	if len(stocks) != 0 {
		fmt.Printf("updating product stock... ")
		err, errApi := s.UpdateStock(stocks)
		if err != nil {
			fmt.Println("error.")
			return err, nil
		}
		if errApi != nil {
			fmt.Println("error.")
			return nil, errApi
		}
		fmt.Println("ok.")
	} else {
		fmt.Println("nothing to update.")
	}

	s.LogFd.Close()

	return nil, nil
}
