package shoper

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	t "time"
)

type Session struct {
	URL   string
	login string
	pass  string

	Token      string `json:"access_token"`
	TokenExpIn int    `json:"expires_in"`
	TokenExp   t.Time

	LogFd *os.File
	log   *log.Logger
}

func NewSession(URL, login, pass string) (*Session, error) {
	var err error
	s := &Session{URL: URL, login: login, pass: pass}

	flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	s.LogFd, err = os.OpenFile("tmp/shoper.log", flags, 0600)
	if err != nil {
		return nil, err
	}

	s.log = log.New(s.LogFd, "", log.LstdFlags)
	s.log.Printf("=== %s session start ===\n", s.URL)

	err = s.tokenFromFile()
	inAweek := t.Now().AddDate(0, 0, 7)
	if err != nil || s.TokenExp.Before(inAweek) {
		err = s.getToken()
		if err != nil {
			s.log.Println(err)
			return nil, err
		}

		s.TokenExp = t.Now().Add(t.Second * t.Duration(s.TokenExpIn))

		err = s.saveToken()
		if err != nil {
			s.log.Println(err)
			return nil, err
		}

		s.log.Println("new token saved to ./tmp/token")
	} else {
		s.log.Println("token read from ./tmp/token")
	}

	s.log.Println("token expiry date: " + s.TokenExp.Format(t.RFC1123Z))

	return s, err
}

func (s *Session) tokenFromFile() error {
	fd, err := os.Open("tmp/token")
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
			return errors.New("too many lines in ./tmp/token file")
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

	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) saveToken() error {
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	fd, err := os.OpenFile("tmp/token", flags, 0600)
	defer fd.Close()
	if err != nil {
		return err
	}

	fmt.Fprintln(fd, s.Token)
	fmt.Fprintln(fd, s.TokenExp.Format(t.RFC1123Z))

	return nil
}
