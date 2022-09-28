package input

import (
	"bufio"
	"os"
	"strings"
)

type ShopT struct {
	Url   string
	Login string
	Pass  string
}

func ParseLoginConf() ([]ShopT, error) {
	var shops []ShopT

	fd, err := os.Open("./etc/login.conf")
	defer fd.Close()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(fd)

	for scanner.Scan() {
		line := strings.Join(strings.Fields(scanner.Text()), " ")
		if strings.HasPrefix(line, "#") {
			continue
		} else if strings.HasPrefix(line, " #") {
			continue
		} else if line == "" || line == " " {
			continue
		}

		fields := strings.Split(line, " ")

		var shop ShopT
		shop.Url = fields[0]
		shop.Login = fields[1]
		shop.Pass = fields[2]

		shops = append(shops, shop)
	}

	return shops, nil
}
