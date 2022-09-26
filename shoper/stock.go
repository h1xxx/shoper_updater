package shoper

import (
	"encoding/json"
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

	NewStock float64
}

type bulkRespT struct {
	Errors bool     `json:"errors"`
	Items  []itemsT `json:"items"`
}

type itemsT struct {
	Code      int        `json:"code"`
	StockPage StockPageT `json:"body"`
}

func (s *Session) getStockPage(page int) (StockPageT, error) {
	var stockPage StockPageT

	url := fmt.Sprintf("%s/webapi/rest/product-stocks/?page=%d&limit=50",
		s.URL, page)

	err := s.callApi(url, "GET", nil, &stockPage)
	if err != nil {
		return stockPage, err
	}

	return stockPage, nil
}

func (s *Session) GetStockList() ([]StockT, error) {
	// get the first page to see the stock and stock pages counts
	stockPage1, err := s.getStockPage(1)
	if err != nil {
		return []StockT{}, err
	}

	pagesCount := stockPage1.Pages
	stockCount, err := strconv.ParseInt(stockPage1.StockCount, 10, 64)
	if err != nil {
		return []StockT{}, errors.New("can't parse stock count")
	}

	// prepare data for json payload
	var bulkDataList [][]byte
	for i := 2; i <= int(pagesCount); i += 25 {
		bulkData, err := getPagesBulk(i, i+24)
		if err != nil {
			msg := "can't prepare bulk data for getting stock info"
			return []StockT{}, errors.New(msg)
		}
		bulkDataList = append(bulkDataList, bulkData)
	}

	// append the stock lists from the api call responses
	var stockList []StockT
	stockList = append(stockList, stockPage1.StockList...)
	for _, data := range bulkDataList {
		var resp bulkRespT
		err = s.callApi(s.URL+"/webapi/rest/bulk/", "PUT", data, &resp)
		if err != nil {
			return []StockT{}, err
		}

		for _, i := range resp.Items {
			stockList = append(stockList, i.StockPage.StockList...)
		}
	}

	// check if all stock IDs were processed

	if len(stockList) != int(stockCount) {
		msg := "len of stock list doesn't match info from API"
		return []StockT{}, errors.New(msg)
	}

	return stockList, nil
}

func getPagesBulk(pageStart, pageEnd int) ([]byte, error) {
	type bulkPageParamsT struct {
		Limit string `json:"limit"`
		Page  string `json:"page"`
	}

	type bulkPageT struct {
		Id     string          `json:"id"`
		Path   string          `json:"path"`
		Params bulkPageParamsT `json:"params"`
		Method string          `json:"method"`
	}

	var data []bulkPageT
	for page := pageStart; page <= pageEnd; page++ {
		var el bulkPageT
		el.Id = fmt.Sprintf("product-stocks-%d", page)
		el.Path = "/webapi/rest/product-stocks"
		el.Method = "GET"

		var params bulkPageParamsT
		params.Limit = "50"
		params.Page = fmt.Sprintf("%d", page)
		el.Params = params

		data = append(data, el)
	}

	postBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return postBody, nil
}

func getProductsBulk(productList []string) ([]byte, error) {
	var data []map[string]string
	for _, id := range productList {
		el := make(map[string]string)
		el["id"] = fmt.Sprintf("product-stocks-%d", id)
		el["path"] = fmt.Sprintf("/webapi/rest/product-stocks/%d", id)
		el["method"] = "GET"
		data = append(data, el)
	}

	postBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return postBody, nil
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

// outputs only products that are in input file and in Stan_mag.txt file
func GetStanMagStock(stocks map[string]StockT, stanMag map[string]float64) map[string]StockT {
	res := make(map[string]StockT)
	for k, v := range stanMag {
		_, exists := stocks[k]
		if exists {
			product := stocks[k]
			product.NewStock = v
			res[k] = product
		}
	}

	return res
}

// outputs only products that need to have stock value updated
func GetUpdateStock(stocks map[string]StockT) (map[string]StockT, error) {
	res := make(map[string]StockT)
	for k, v := range stocks {
		stock, err := strconv.ParseFloat(v.Stock, 64)
		if err != nil {
			msg := "can't parse product stock count"
			return res, errors.New(msg)
		}

		if stock != v.NewStock {
			res[k] = v
		}
	}

	return res, nil
}
