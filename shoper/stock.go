package shoper

import (
	"fmt"
)

type StockPageT struct {
	StockCount string   `json:"count"`
	Pages      int      `json:"pages"`
	CurrPage   int      `json:"page"`
	StockList  []StockT `json:"list"`
}

type StockT struct {
	StockID   string `json:"stock_id"`
	Price     string `json:"price"`
	PriceType string `json:"price_type"`
	Stock     string `json:"stock"`
	Sold      string `json:"sold"`
	Active    string `json:"active"`
	ProductID string `json:"product_id"`
	Ean       string `json:"ean"`
}

func (s *Session) GetStock(page int) (StockPageT, error) {
	var stockPage StockPageT

	url := fmt.Sprintf("%s/webapi/rest/product-stocks/?page=%d&limit=50",
		s.URL, page)

	err := s.callApi(url, "GET", "", &stockPage)
	if err != nil {
		return stockPage, err
	}

	return stockPage, nil
}
