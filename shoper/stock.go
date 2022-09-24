package shoper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

	client := &http.Client{}
	url := fmt.Sprintf("%s/webapi/rest/product-stocks/?page=%d&limit=50",
		s.URL, page)
	req, err := http.NewRequest("GET", url, nil)
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

	s.apiCallsLeft, err = getApiCallsLeft(resp.Header)
	if err != nil {
		return stockPage, err
	}

	if s.apiCallsLeft < 5 {
		// todo: implement sleep
		return stockPage, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&stockPage)
	if err != nil {
		return stockPage, err
	}

	return stockPage, nil
}

func getApiCallsLeft(headers http.Header) (int, error) {
	callsStr, callsExists := headers["X-Shop-Api-Calls"]
	limitStr, limitExists := headers["X-Shop-Api-Limit"]
	if !(callsExists && limitExists) {
		return 0, errors.New("no 'X-Shop-Api' headers in response")
	}

	calls, err := strconv.ParseInt(callsStr[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid X-Shop-Api-Calls header")
	}

	limit, err := strconv.ParseInt(limitStr[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid X-Shop-Api-Limit header")
	}

	apiCallsLeft := int(limit - calls)

	return apiCallsLeft, nil
}
