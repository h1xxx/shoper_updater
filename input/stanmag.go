package input

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
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

func AddSets(stanMag map[string]float64, sets []ProductSetT) {
	for _, set := range sets {
		stanMag[set.SetCode] = calcSetStock(stanMag, set)
	}
}

func calcSetStock(stanMag map[string]float64, set ProductSetT) float64 {
	var setStock float64

	_, setInStanmag := stanMag[set.SetCode]
	if setInStanmag {
		msg := "WARNING! Product set %s already exists in Stan_mag.txt,"
		msg += " and its stock value will be recalculated.\n"
		fmt.Printf(msg, set.SetCode)
	}

	for part, val := range set.Parts {
		stock, exists := stanMag[part]
		if exists {
			if stock == 0 {
				return 0
			}
			if setStock == 0 {
				setStock = math.Floor(stock / val)
			} else {
				setStock = math.Min(math.Floor(stock/val),
					setStock)
			}
		} else {
			msg := "WARNING! Defined part %s from set %s does not"
			msg += "exist in Stan_mag.txt\n"
			fmt.Printf(msg, part, set.SetCode)
			return 0
		}
	}
	return setStock
}
