package input

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func ParseStanMag(file string) map[string]float64 {
	stanMag := make(map[string]float64)
	f, err := os.Open(file)
	errExit(err, "can't open file: "+file)
	scanner := bufio.NewScanner(f)

	var i int
	for scanner.Scan() {
		i += 1
		line := scanner.Text()
		fields := strings.Split(line, "\t")

		w := "line needs to have 3 fields, ignoring"
		msg := fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		if len(fields) != 3 {
			warn(msg)
		}

		code := fields[0]
		quantity := strings.Replace(fields[1], ",", ".", -1)

		if code == "" || code == "SYMBOL" {
			continue
		}

		w = "1st field too short, ignoring"
		msg = fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		if len(code) < 4 {
			warn(msg)
			continue
		}

		quantityFl, err := strconv.ParseFloat(quantity, 64)

		w = "2nd field not a float number, ignoring"
		msg = fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		if err != nil || !strings.Contains(quantity, ".") {
			warn(msg)
			continue
		}

		w = "product code already exists, ignoring"
		msg = fmt.Sprintf("Data_mag.txt:%d %s\n%s", i, w, line)
		_, exists := stanMag[code]
		if exists {
			warn(msg)
			continue
		}

		stanMag[code] = quantityFl
	}

	return stanMag
}

func warn(msg string) {
	//log.Println("WARNING: " + msg)
	return
}

func errExit(err error, msg string) {
	if err != nil {
		log.Println("\n * " + msg)
		log.Fatal(err)
	}
}
