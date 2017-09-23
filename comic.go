package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type comic struct {
	URL   string
	Match string
	Last  string
}

func parseComics(content []byte) (comics []comic) {
	scanner := bufio.NewScanner(bytes.NewReader(content))

	var c comic
	for scanner.Scan() {
		if scanner.Text() == "" {
			break
		}
		c.URL = scanner.Text()
		scanner.Scan()
		c.Match = scanner.Text()
		scanner.Scan()
		c.Last = scanner.Text()
		comics = append(comics, c)
	}

	return comics
}

func openComic(url string, browser string) {
	var cmdStr string
	var args string
	if runtime.GOOS == "windows" {
		cmdStr = "cmd"
		args = "/c start "
	} else {
		return
	}

	cmd := exec.Command(cmdStr, args+browser+" "+url)
	err := cmd.Run()
	if err != nil {
		log.Printf("Unable to execute command: %s\n", err)
	}

}

func checker(comics []comic, comicsPipe chan comic, browser string, throttle chan bool) {
	var wg sync.WaitGroup

	for _, c := range comics {
		wg.Add(1)
		throttle <- true
		go checkComic(c, comicsPipe, browser, &wg, throttle)
	}
	wg.Wait()
	close(comicsPipe)
}

//Add check if no match found
func checkComic(c comic, comicsPipe chan comic, browser string, wg *sync.WaitGroup, throttle chan bool) {
	resp, err := http.Get(c.URL)
	if err != nil {
		log.Printf("Unable to fetch %s: %s\n", c.URL, err)

	} else {
		defer resp.Body.Close()

		found := false

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, c.Match) {
				found = true
				if line != c.Last {
					openComic(c.URL, browser)
					c.Last = line
					break
				}
			}
		}

		if found == false {
			fmt.Printf("Didn't find %s\n", c.URL)
		}
		comicsPipe <- c
	}
	<-throttle
	wg.Done()
}

func main() {

	comicsFile := "comics.txt"
	maxConcurrentRequests := 10
	browser := "chrome"
	newline := []byte{'\r', '\n'}
	if runtime.GOOS != "windows" {
		newline = []byte{'\n'}
	}

	content, err := ioutil.ReadFile(comicsFile)
	if err != nil {
		log.Fatalf("Unable to open comics file: %s\n", err)
	}
	comics := parseComics(content)

	comicsPipe := make(chan comic)

	now := time.Now()
	throttle := make(chan bool, maxConcurrentRequests)
	go checker(comics, comicsPipe, browser, throttle)

	var newContent []byte
	for c := range comicsPipe {
		newContent = append(newContent, []byte(c.URL)...)
		newContent = append(newContent, newline...)
		newContent = append(newContent, []byte(c.Match)...)
		newContent = append(newContent, newline...)
		newContent = append(newContent, []byte(c.Last)...)
		newContent = append(newContent, newline...)
	}
	err = ioutil.WriteFile(comicsFile, newContent, 644)
	if err != nil {
		log.Fatalf("Unable to save comics file\n")
	}
	fmt.Printf("Time: %s\n", time.Since(now))
}
