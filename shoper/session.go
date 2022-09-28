package shoper

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"strings"
	t "time"
)

type Session struct {
	URL   string
	login string
	pass  string

	Domain     string
	Token      string `json:"access_token"`
	TokenExpIn int    `json:"expires_in"`
	TokenExp   t.Time

	LogFd *os.File
	log   *log.Logger

	apiCallsLeft int
}

func NewSession(URL, login, pass string) (*Session, error) {
	var err error
	s := &Session{URL: URL, login: login, pass: pass}
	s.Domain = strings.Trim(fp.Base(URL), "www.")

	flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	s.LogFd, err = os.OpenFile("log/shoper.log", flags, 0600)
	if err != nil {
		return nil, err
	}

	s.log = log.New(s.LogFd, "", log.LstdFlags)
	s.log.Printf("=== %s session start ===\n", s.URL)

	err = s.tokenFromFile(s.Domain)
	inAweek := t.Now().AddDate(0, 0, 7)
	if err != nil || s.TokenExp.Before(inAweek) {
		err = s.getToken()
		if err != nil {
			s.log.Println(err)
			return nil, err
		}

		s.TokenExp = t.Now().Add(t.Second * t.Duration(s.TokenExpIn))

		err = s.saveToken(s.Domain)
		if err != nil {
			s.log.Println(err)
			return nil, err
		}

		s.log.Println("new token saved to ./log/token_" + s.Domain)
	} else {
		s.log.Println("token read from ./log/token_" + s.Domain)
	}

	s.log.Println("token expiry date: " + s.TokenExp.Format(t.RFC1123Z))

	return s, err
}

func (s *Session) tokenFromFile(domain string) error {
	fd, err := os.Open("log/token_" + domain)
	defer fd.Close()
	if err != nil {
		return err
	}

	var i int
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		i += 1
		line := scanner.Text()
		switch i {
		case 1:
			s.Token = line
		case 2:
			s.TokenExp, err = t.Parse(t.RFC1123Z, line)
			if err != nil {
				return err
			}
		default:
			msg := "too many lines in ./log/token_" + domain
			return errors.New(msg)
		}
	}

	return nil
}

func (s *Session) getToken() error {
	client := &http.Client{}
	req, err := http.NewRequest("POST", s.URL+"/webapi/rest/auth", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "curl/7.77.0")
	req.SetBasicAuth(s.login, s.pass)
	req.Close = true

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("error getting token: %d - %s",
			resp.StatusCode, ERRCODE[resp.StatusCode])
		return errors.New(msg)
	}

	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) saveToken(domain string) error {
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	fd, err := os.OpenFile("log/token_"+domain, flags, 0600)
	defer fd.Close()
	if err != nil {
		return err
	}

	fmt.Fprintln(fd, s.Token)
	fmt.Fprintln(fd, s.TokenExp.Format(t.RFC1123Z))

	return nil
}
