package input

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func ParseStanMag(file string) (map[string]float64, int, error) {
	stanMag := make(map[string]float64)
	var errCount, emptyProd int

	flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	logFd, err := os.OpenFile("log/stan_mag.log", flags, 0600)
	defer logFd.Close()
	if err != nil {
		return nil, 0, err
	}

	l := log.New(logFd, "", log.LstdFlags)
	l.Println("=== parsing start ===")

	fd, err := os.Open(file)
	defer fd.Close()
	if err != nil {
		return nil, errCount, err
	}
	scanner := bufio.NewScanner(fd)

	var i int
	for scanner.Scan() {
		i += 1
		line := scanner.Text()
		fields := strings.Split(line, "\t")

		w := "line needs to have 3 fields"
		msg := fmt.Sprintf("line:%d %s\t'%s'", i, w, line)
		if len(fields) != 3 {
			errCount++
			l.Println(msg)
		}

		code := fields[0]
		quantity := strings.Replace(fields[1], ",", ".", -1)

		if code == "SYMBOL" {
			continue
		}

		if code == "" {
			emptyProd++
			errCount++
			continue
		}

		w = "1st field too short"
		msg = fmt.Sprintf("line:%d %s\t\t'%s'", i, w, line)
		if len(code) < 4 {
			errCount++
			l.Println(msg)
			continue
		}

		quantityFl, err := strconv.ParseFloat(quantity, 64)

		w = "2nd field not a float number"
		msg = fmt.Sprintf("line:%d %s\t'%s'", i, w, line)
		if err != nil || !strings.Contains(quantity, ".") {
			errCount++
			l.Println(msg)
			continue
		}

		w = "2nd field not a positive number"
		msg = fmt.Sprintf("line:%d %s\t'%s'", i, w, line)
		if quantityFl < 0 {
			errCount++
			l.Println(msg)
			continue
		}

		w = "product code already exists"
		msg = fmt.Sprintf("line:%d %s\t'%s'", i, w, line)
		_, exists := stanMag[code]
		if exists {
			errCount++
			l.Println(msg)
			continue
		}

		stanMag[code] = quantityFl
	}

	l.Printf("errors: %d\n", errCount)
	l.Printf("lines with empty product: %d\n", emptyProd)

	if len(stanMag) == 0 {
		return nil, 0, errors.New("no valid entries")
	}

	return stanMag, errCount, nil
}
