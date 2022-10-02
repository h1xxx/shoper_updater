package shoper

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// structs for normal bulk response during api calls
type bulkRespPutT struct {
	Errors bool        `json:"errors"`
	Items  []itemsPutT `json:"items"`
}
type itemsPutT struct {
	Code int    `json:"code"`
	Body int    `json:"body"`
	ID   string `json:"id"`
}

// structs for error bulk response during api calls
type bulkErrRespPutT struct {
	Errors bool           `json:"errors"`
	Items  []itemsErrPutT `json:"items"`
}
type itemsErrPutT struct {
	Code int         `json:"code"`
	Body bodyPutErrT `json:"body"`
	ID   string      `json:"id"`
}
type bodyPutErrT struct {
	Error string `json:"error"`
	Desc  string `json:"error_description"`
}

// first returned err is for general errors, second one for errors from api
func (s *Session) UpdateStock(stocks map[string]StockT) (error, error) {
	// prepare a list of stocks to update and sort it by stock_id
	var stockList []StockT
	for _, stock := range stocks {
		stockList = append(stockList, stock)
	}
	sort.Slice(stockList, func(i, j int) bool {
		return stockList[i].ProductCode <= stockList[j].ProductCode
	})

	// create a stock_id => product_code map for error logging
	idToCode := make(map[string]string)

	// log changes to be made
	s.log.Println("starting product stock update")
	for _, stock := range stockList {
		s.log.Printf("%6s : %-14s %6s => %d.0\n", stock.StockID,
			stock.ProductCode, stock.Stock, stock.NewStock)
		idToCode[stock.StockID] = stock.ProductCode
	}

	// prepare data for json payload
	var bulkDataList [][]byte
	var reqCount int
	for i := 0; i <= len(stockList); i += 25 {
		bulkData, count, err := makeStocksPutData(stockList, i,
			min(i+24, len(stocks)-1))
		if err != nil {
			msg := "can't prepare bulk data for updating stock info"
			return errors.New("UpdateStock:: " + msg), nil
		}
		bulkDataList = append(bulkDataList, bulkData)
		reqCount += count
	}

	if len(stocks) != reqCount {
		msg := "count of products to update doesn't match the count "
		msg += "of requests.\ntry again or find a developer to solve "
		msg += "this"
		s.log.Println("error when preparing data. no update submitted")
		return errors.New("UpdateStock:: " + msg), nil
	}

	// make the api calls to update stock info
	for i, data := range bulkDataList {
		var resp bulkRespPutT
		var respErr bulkErrRespPutT
		err := s.callApi(s.URL+"/webapi/rest/bulk/", "PUT", data,
			&resp, &respErr)

		var msgCum string

		for _, i := range respErr.Items {
			id := strings.TrimPrefix(i.ID, "product-stocks-")
			if i.Code != 200 {
				msg := fmt.Sprintf("%d %s | %s (%s) | %s",
					i.Code, ERRCODE[i.Code],
					id, idToCode[id], i.Body.Desc)
				s.log.Println(msg)

				fmtStr := "%d %s for %s (%s): \t%s\n"
				msgCum += fmt.Sprintf(fmtStr,
					i.Code, ERRCODE[i.Code],
					id, idToCode[id], i.Body.Desc)
			}
		}

		if resp.Errors || msgCum != "" {
			return nil, errors.New(msgCum)
		}

		if err != nil {
			msg := fmt.Sprintf("UpdateStock:: err on batch %d", i)
			s.log.Printf("%s %s\n", msg, err)
			return fmt.Errorf("UpdateStock:: %w", err), nil
		}
	}

	s.log.Println("update successful")

	return nil, nil
}

func makeStocksPutData(stockList []StockT, start, end int) ([]byte, int, error) {
	if len(stockList) == 0 {
		return nil, 0, nil
	}
	type bulkPageBodyT struct {
		Stock string `json:"stock"`
	}

	type bulkPageT struct {
		Id     string        `json:"id"`
		Path   string        `json:"path"`
		Body   bulkPageBodyT `json:"body"`
		Method string        `json:"method"`
	}

	var data []bulkPageT
	var reqCount int

	for i := start; i <= end; i++ {
		stock := stockList[i]
		id := stock.StockID
		if id == "" {
			continue
		}

		var el bulkPageT
		el.Id = fmt.Sprintf("product-stocks-%s", id)
		el.Path = fmt.Sprintf("/webapi/rest/product-stocks/%s", id)
		el.Method = "PUT"

		var body bulkPageBodyT
		body.Stock = fmt.Sprintf("%d.0", stock.NewStock)
		el.Body = body

		data = append(data, el)
		reqCount += 1
	}

	postBody, err := json.Marshal(data)
	if err != nil {
		return nil, reqCount, err
	}

	return postBody, reqCount, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
