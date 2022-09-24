package shoper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (s *Session) callApi(url, method, data string, res interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "curl/7.77.0")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Close = true

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("error making api call: %d", resp.StatusCode)
		return errors.New(msg)
	}

	s.apiCallsLeft, err = getApiCallsLeft(resp.Header)
	if err != nil {
		return err
	}

	if s.apiCallsLeft < 5 {
		// todo: implement sleep
		return nil
	}

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
