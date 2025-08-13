package main

import "log"

func main() {
	if err := Run2Agents1User(); err != nil {
		log.Fatal(err)
	}
}
