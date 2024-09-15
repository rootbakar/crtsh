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
		Timeout: 25 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf(BASE_URL, domain))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var jsonData []map[string]string
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		return
	}

	for _, entry := range jsonData {
		nameValue := entry["name_value"]
		subnames := strings.Split(nameValue, "\n")
		for _, subname := range subnames {
			if strings.Contains(subname, "*") {
				if !wildcardsubdomains[subname] {
					wildcardsubdomains[subname] = true
				}
			} else {
				if !subdomains[subname] {
					subdomains[subname] = true
				}
			}
		}
	}
}

func main() {
	domain, recursive, wildcard := parseArgs()

	crtsh(domain)

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
