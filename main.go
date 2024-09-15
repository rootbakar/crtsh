package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const BASE_URL = "https://crt.sh/?q=%s&output=json"

var subdomains = make(map[string]bool)
var wildcardsubdomains = make(map[string]bool)

func parserError(errmsg string) {
	fmt.Println("Usage: go run main.go [Options] use -h for help")
	fmt.Println("Error: " + errmsg)
	os.Exit(1)
}

func parseArgs() (string, bool, bool) {
	domain := flag.String("d", "", "Specify Target Domain to get subdomains from crt.sh")
	recursive := flag.Bool("r", false, "Do recursive search for subdomains")
	wildcard := flag.Bool("w", false, "Include wildcard in output")

	flag.Parse()

	if *domain == "" {
		parserError("Domain is required")
	}

	return *domain, *recursive, *wildcard
}

func crtsh(domain string) {
	client := http.Client{
		Timeout: 60 * time.Second, // Tingkatkan timeout menjadi 60 detik
	}

	retries := 3 // Retry up to 3 times
	for retries > 0 {
		resp, err := client.Get(fmt.Sprintf(BASE_URL, domain))
		if err != nil {
			fmt.Println("Error making request:", err)
			retries-- // Kurangi jumlah retry
			time.Sleep(2 * time.Second) // Tunggu 2 detik sebelum mencoba ulang
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Non-OK HTTP status:", resp.StatusCode)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		if len(body) == 0 {
			fmt.Println("No data found for domain:", domain)
			return
		}

		var jsonData []map[string]interface{}
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		for _, entry := range jsonData {
			nameValue, ok := entry["name_value"].(string)
			if !ok {
				continue
			}
			subnames := strings.Split(nameValue, "\n")
			for _, subname := range subnames {
				if strings.Contains(subname, "*") {
					wildcardsubdomains[subname] = true
				} else {
					subdomains[subname] = true
				}
			}
		}
		break // Keluar dari loop jika berhasil
	}
}

func main() {
	domain, recursive, wildcard := parseArgs()

	crtsh(domain)

	if len(subdomains) == 0 && len(wildcardsubdomains) == 0 {
		fmt.Println("No subdomains found for domain:", domain)
		return
	}

	for subdomain := range subdomains {
		fmt.Println(subdomain)
	}

	if recursive {
		for wildcardsubdomain := range wildcardsubdomains {
			wildcardsubdomain = strings.ReplaceAll(wildcardsubdomain, "*.", "%25.")
			crtsh(wildcardsubdomain)
		}
	}

	if wildcard {
		for wildcardsubdomain := range wildcardsubdomains {
			fmt.Println(wildcardsubdomain)
		}
	}
}
