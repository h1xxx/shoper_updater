package shoper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (s *Session) callApi(url, method string, data []byte, res interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "curl/7.77.0")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Close = true

	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("error making api call: %d", resp.StatusCode)
		return errors.New(msg)
	}

	s.apiCallsLeft, err = getApiCallsLeft(resp.Header)
	if err != nil {
		return err
	}
	slowDown(s.apiCallsLeft)

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return err
	}

	return nil
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

func slowDown(callsLeft int) {
	switch {
	case callsLeft <= 1:
		time.Sleep(30 * time.Second)
	case callsLeft <= 3:
		time.Sleep(10 * time.Second)
	case callsLeft <= 5:
		time.Sleep(500 * time.Millisecond)
	case callsLeft <= 8:
		time.Sleep(200 * time.Millisecond)
	}
}
