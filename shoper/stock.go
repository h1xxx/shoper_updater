package shoper

import (
	"errors"
	"fmt"
	"strconv"
)

type StockPageT struct {
	StockCount string   `json:"count"`
	Pages      int      `json:"pages"`
	CurrPage   int      `json:"page"`
	StockList  []StockT `json:"list"`
}

type StockT struct {
	StockID     string `json:"stock_id"`
	Price       string `json:"price"`
	PriceType   string `json:"price_type"`
	Stock       string `json:"stock"`
	Sold        string `json:"sold"`
	Active      string `json:"active"`
	ProductID   string `json:"product_id"`
	ProductCode string `json:"code"`
	EAN         string `json:"ean"`
}

func (s *Session) getStockPage(page int) (StockPageT, error) {
	var stockPage StockPageT

	url := fmt.Sprintf("%s/webapi/rest/product-stocks/?page=%d&limit=50",
		s.URL, page)

	err := s.callApi(url, "GET", "", &stockPage)
	if err != nil {
		return stockPage, err
	}

	return stockPage, nil
}

func (s *Session) GetStockList() ([]StockT, error) {
	var stockList []StockT
	var stockPage StockPageT

	// getting the first page shows how many pages there are
	stockFirstPage, err := s.getStockPage(1)
	if err != nil {
		return stockList, err
	}
	stockList = append(stockList, stockFirstPage.StockList...)
	stockCount, err := strconv.ParseInt(stockFirstPage.StockCount, 10, 64)
	if err != nil {
		return stockList, errors.New("can't parse stock count")
	}

	// iterate all remaining pages
	for page := 2; page <= stockFirstPage.Pages; page++ {
		stockPage, err = s.getStockPage(page)
		if err != nil {
			return stockList, err
		}
		stockList = append(stockList, stockPage.StockList...)
	}

	// check if all stock IDs were processed
	if len(stockList) != int(stockCount) {
		msg := "len of stock list doesn't match info from API"
		return stockList, errors.New(msg)
	}

	return stockList, nil
}

func GetStockMap(stockList []StockT) (map[string]StockT, error) {
	stocks := make(map[string]StockT)
	for _, s := range stockList {
		if s.ProductCode == "" {
			// todo: make some threshold on count of this case
			continue
		}

		_, exists := stocks[s.ProductCode]
		if exists {
			fmt.Println("product code already exists:", s)
			continue
		}

		stocks[s.ProductCode] = s
	}
	return stocks, nil
}
