package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	timeoutInSeconds = 5
	concurrentLimit  = 10
)

func addDoubleSlash(urlStr string) string {
	if strings.HasPrefix(urlStr, "https://") || strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https:////") {
		parts := strings.Split(urlStr, "/")

		if len(parts) > 3 {
			parts[2] += "/"
			newURL := strings.Join(parts, "/")
			return newURL + "/"
		} else {
			return urlStr + "/"
		}
	} else {
		return urlStr + "/"
	}
}

func checkPatternsInHTML(urlStr string, patterns []string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := http.Client{
		Timeout: timeoutInSeconds * time.Second,
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			htmlLine := scanner.Text()
			for _, pattern := range patterns {
				if strings.Contains(htmlLine, pattern) {
					fmt.Printf("\033[32m[Reflected] %s\033[0m\n", urlStr)
					return
				}
			}
		}
	}
	fmt.Println("[Not reflected]", urlStr)
}

func main() {
	listFilePath := flag.String("l", "", "Path to the input text file with URLs")
	flag.Parse()

	if *listFilePath == "" {
		fmt.Println("Please provide the path to the input text file using -l")
		return
	}

	file, err := os.Open(*listFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var urls []string
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	var newURLs []string
	for _, url := range urls {
		newURLs = append(newURLs, addDoubleSlash(strings.TrimSpace(url)))
	}

	patternsToCheck := []string{
		"/url",
		"/%0a/url",
		"/%0d/url",
		"/<>/url",
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrentLimit)

	for _, newURL := range newURLs {
		wg.Add(1)
		go func(urlStr string) {
			semaphore <- struct{}{}
			checkPatternsInHTML(urlStr, patternsToCheck, &wg)
			<-semaphore
		}(newURL)
	}

	wg.Wait()
}
