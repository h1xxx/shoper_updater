package input

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseStanMag(file string) (map[string]float64, int, error) {
	stanMag := make(map[string]float64)
	var errCount int

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

		w := "line needs to have 3 fields, ignoring"
		msg := fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		if len(fields) != 3 {
			errCount++
			warn(msg)
		}

		code := fields[0]
		quantity := strings.Replace(fields[1], ",", ".", -1)

		if code == "SYMBOL" {
			continue
		}

		if code == "" {
			errCount++
			continue
		}

		w = "1st field too short, ignoring"
		msg = fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		if len(code) < 4 {
			errCount++
			warn(msg)
			continue
		}

		quantityFl, err := strconv.ParseFloat(quantity, 64)

		w = "2nd field not a float number, ignoring"
		msg = fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		if err != nil || !strings.Contains(quantity, ".") {
			errCount++
			warn(msg)
			continue
		}

		w = "product code already exists, ignoring"
		msg = fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		_, exists := stanMag[code]
		if exists {
			errCount++
			warn(msg)
			continue
		}

		stanMag[code] = quantityFl
	}

	return stanMag, errCount, nil
}

func warn(msg string) {
	//log.Println("WARNING: " + msg)
	return
}
