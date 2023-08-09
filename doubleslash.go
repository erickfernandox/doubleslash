package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	timeoutInSeconds = 5
)

func addDoubleSlash(urlStr string) string {
	if strings.HasPrefix(urlStr, "https://") || strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https:////") {
		parts := strings.Split(urlStr, "/")

		if len(parts) > 3 {
			parts[2] += "/"
			newURL := strings.Join(parts, "/")
			return newURL
		} else {
			return urlStr
		}
	} else {
		return urlStr
	}
}

func checkDoubleSlashInHTML(urlStr string) bool {
	client := http.Client{
		Timeout: timeoutInSeconds * time.Second,
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			htmlLine := scanner.Text()
			parsedURL, _ := url.Parse(urlStr)
			if parsedURL.Path != "//" && parsedURL.Path != "" {
				if strings.Contains(htmlLine, "=\""+parsedURL.Path) {
					return true
				} else if strings.Contains(htmlLine, "=\"https:"+parsedURL.Path) {
					return true
				}
			}
		}
	}

	return false
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

	for _, newURL := range newURLs {
		if checkDoubleSlashInHTML(newURL) {
			fmt.Printf("\033[32m[Reflected] %s\033[0m\n", newURL)
		} else {
			fmt.Println("[Not reflected]", newURL)
		}
	}
}
