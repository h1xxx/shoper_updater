package shoper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

func (s *Session) GetStock() (StockPageT, error) {
	var stockPage StockPageT

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		s.URL+"/webapi/rest/product-stocks/?limit=50", nil)
	if err != nil {
		return stockPage, err
	}
	req.Header.Set("User-Agent", "curl/7.77.0")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Close = true

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return stockPage, err
	}
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("error getting stock page: %d",
			resp.StatusCode)
		return stockPage, errors.New(msg)
	}

	err = json.NewDecoder(resp.Body).Decode(&stockPage)
	if err != nil {
		return stockPage, err
	}

	return stockPage, nil
}
