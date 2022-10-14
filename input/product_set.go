package input

import (
	"bufio"
	//"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// parts contains a map of product_set => quantity
type ProductSetT struct {
	Name    string
	SetCode string
	Parts   map[string]float64
}

func ParseProductSets(file string) ([]ProductSetT, int, error) {
	var sets []ProductSetT
	var errCount int

	flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	logFd, err := os.OpenFile("log/product_set.log", flags, 0600)
	defer logFd.Close()
	if err != nil {
		return sets, errCount, err
	}

	l := log.New(logFd, "", log.LstdFlags)
	l.Println("=== parsing start ===")

	fd, err := os.Open(file)
	defer fd.Close()
	if err != nil {
		return sets, errCount, err
	}

	// deal with the optional UTF-8 BOM added in Windows
	var bom [3]byte
	_, err = io.ReadFull(fd, bom[:])
	if err != nil {
		return sets, errCount, err
	}
	if bom[0] != 0xef || bom[1] != 0xbb || bom[2] != 0xbf {
		// not a BOM - go back to the beginning of the file
		_, err = fd.Seek(0, 0)
		if err != nil {
			return sets, errCount, err
		}
	}

	scanner := bufio.NewScanner(fd)

	var i int
	var skipSet bool
	var set ProductSetT
	set.Parts = make(map[string]float64)

	var errAny error
	errMsg := "WARNING! Errors in etc/product_set.conf. Check the logs."

	for scanner.Scan() {
		i += 1
		line := trim(scanner.Text())
		fields := strings.Split(line, "#")
		line = trim(fields[0])

		switch {
		case line == "" && set.Name == "":
			continue

		case line == "":
			if !skipSet && setOK(set, l) {
				l.Printf("adding product set %s (parts_%v)\n",
					set.SetCode, set.Parts)
				sets = append(sets, set)
			} else {
				l.Printf("ignoring product set %s (parts_%v)\n",
					set.SetCode, set.Parts)
				errAny = fmt.Errorf(errMsg)
				errCount++
			}
			set = ProductSetT{}
			set.Parts = make(map[string]float64)
			skipSet = false
			continue

		case skipSet:
			continue
		}

		fields = strings.Split(line, "=")
		if len(fields) != 2 {
			l.Printf("ERROR: line %d; too may '=' chars\n", i)
			errAny = fmt.Errorf(errMsg)
			errCount++
			skipSet = true
			continue
		}
		field := trim(fields[0])
		val := trim(fields[1])

		switch {
		case field == "name":
			set.Name = val
			continue

		case field == "set_code":
			if len(val) <= 3 {
				l.Printf("ERROR: line %d; code too short\n", i)
				errAny = fmt.Errorf(errMsg)
				errCount++
				skipSet = true
				continue
			}
			set.SetCode = val
			continue

		case field == "part":
			fields := strings.Split(val, "::")
			if len(fields) != 2 {
				l.Printf("ERROR: line %d; incorrect value\n", i)
				errAny = fmt.Errorf(errMsg)
				errCount++
				skipSet = true
				continue
			}
			partCode := trim(fields[0])
			if len(partCode) <= 3 {
				l.Printf("ERROR: line %d; code too short\n", i)
				errAny = fmt.Errorf(errMsg)
				errCount++
				skipSet = true
				continue
			}
			quant, err := strconv.ParseInt(trim(fields[1]), 10, 64)
			if err != nil {
				l.Printf("ERROR: line %d; incorrect int\n", i)
				errAny = fmt.Errorf(errMsg)
				errCount++
				skipSet = true
				continue
			}
			set.Parts[partCode] = float64(quant)
			continue
		}

		l.Printf("line %d unknown\n", i)
		errAny = fmt.Errorf(errMsg)
		errCount++
	}

	return sets, errCount, errAny
}

func trim(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.Trim(s, " ")

	return s
}

func setOK(set ProductSetT, l *log.Logger) bool {
	if len(set.Parts) <= 1 {
		l.Printf("ERROR: %s has only one part\n", set.SetCode)
		return false
	}
	if set.Name != "" && set.SetCode != "" {
		return true
	}
	return false
}
