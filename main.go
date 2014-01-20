package main

import (
	"fmt"
  "github.com/neocortical/oaklogger/scraper"
)

func main() {
	fmt.Printf("Starting scraper...\n")
	scraper.Scrape()
	fmt.Printf("Interrupt detected ... exiting.\n")
}
