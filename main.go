package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/patrickmn/go-cache"
)

// instance is made to pass the cache around and have some methods.
type instance struct {
	cache *cache.Cache
}

func main() {
	linkCache := cache.New(10*time.Minute, 10*time.Minute)
	inst := &instance{cache: linkCache}
	httpServer := http.Server{
		Addr:              ":7755",
		ReadTimeout:       time.Second * 15,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 30,
		IdleTimeout:       time.Minute * 30,
	}

	http.HandleFunc("/", inst.getIndex)
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func (inst *instance) getIndex(w http.ResponseWriter, r *http.Request) {
	downloadURL, err := inst.getLatestVersionFromCache()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, downloadURL, http.StatusTemporaryRedirect)
}

func (inst *instance) getLatestVersionFromCache() (string, error) {
	url, exists := inst.cache.Get("url")
	if !exists {
		latestURL, err := getLatestVersion()
		if err != nil {
			return "", err
		}
		inst.cache.Set("url", latestURL, cache.DefaultExpiration)
		url = latestURL
	}
	downloadLink, ok := url.(string)
	if !ok {
		log.Println("Unable to convert download link from cache to string.")
		return "", fmt.Errorf("could not find latest version")
	}
	return downloadLink, nil
}

func getLatestVersion() (string, error) {
	resp, err := soup.Get("https://www.minecraft.net/en-us/download/server/bedrock")
	if err != nil {
		log.Println("Unable to get minecraft page.")
		return "", err
	}
	latestVersionURL := ""
	doc := soup.HTMLParse(resp)
	links := doc.FindAll("a", "class", "downloadlink")
	for _, link := range links {
		dataPlatform, exists := link.Attrs()["data-platform"]
		if !exists {
			continue
		}
		if dataPlatform == "serverBedrockLinux" {
			latestVersionURL = link.Attrs()["href"]
			break
		}
	}
	if latestVersionURL == "" {
		log.Println("Could not find latest version from parsed download page.")
		return latestVersionURL, fmt.Errorf("could not find latest version")
	}
	return latestVersionURL, nil
}
