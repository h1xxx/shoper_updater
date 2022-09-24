package main

import (
	"fmt"
	"log"
	"os"

	in "shoper_updater/input"
	sh "shoper_updater/shoper"
)

func main() {
	os.MkdirAll("tmp/", 0700)

	stanMag := in.ParseStanMag("data/Stan_mag.txt")
	fmt.Println("Stan_mag.txt products:", len(stanMag))

	if len(os.Args) > 20 {
		log.Fatalln("Usage: main <searchterm>")
	}

	url := os.Getenv("SHOPER_URL")
	login := os.Getenv("SHOPER_LOGIN")
	pass := os.Getenv("SHOPER_PASS")

	s, err := sh.NewSession(url, login, pass)
	if err != nil {
		log.Panicln(err)
	}

	s.LogFd.Close()
}

func errExit(err error, msg string) {
	if err != nil {
		log.Println("\n * " + msg)
		log.Fatal(err)
	}
}
