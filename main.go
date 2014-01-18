package main

import (
	"fmt"
  "github.com/neocortical/oaklogger/scraper"
)

func main() {
	fmt.Printf("Starting scraper...\n")
	go scraper.Scrape()
	fmt.Printf("Interrupt detected ... exiting.\n")
}
