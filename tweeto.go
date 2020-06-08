package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"mvdan.cc/xurls/v2"
)

// OEmbed is the response from Twitter
type OEmbed struct {
	EmbedType    string `json:"type"`
	URL          string
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	HTML         string
	Width        int
	Height       int
	CacheAge     string `json:"cache_age"`
	ProviderName string `json:"provider_name"`
	ProviderURL  string `json:"provider_url"`
	Version      string
}

func getURLFromBody(s string) (string, error) {
	rxStrict := xurls.Strict()
	u := rxStrict.FindString(s)
	if u == "" {
		return "", fmt.Errorf("No url found")
	}
	return u, nil
}

func getPreviewFromURL(uri string) (string, error) {
	if strings.Contains(uri, "twitter.com") {
		// returns a JSON oEmbed response.
		apiEndpoint := fmt.Sprintf("https://publish.twitter.com/oembed?url=%s", uri)
		client := http.Client{Timeout: time.Second * 2}
		resp, err := client.Get(apiEndpoint)
		if err != nil {
			fmt.Println(err)
			return "", fmt.Errorf("url preview cannot connect")
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("url preview cannot read response: %w", err)
			}

			var embed OEmbed
			if err = json.Unmarshal(data, &embed); err != nil {
				return "", fmt.Errorf("url preview cannot read response: %w", err)

			}
			// Remove js and html
			policy := bluemonday.StrictPolicy()
			policy.RequireParseableURLs(true)
			policy.AllowRelativeURLs(false)
			policy.AllowURLSchemes("http", "https")
			html := policy.Sanitize(embed.HTML)
			// hopping keybase will preview that. Edit: nope
			html = strings.ReplaceAll(html, "pic.twitter.com", "https://pic.twitter.com")

			return "\n" + html, nil
		}
		return "", fmt.Errorf("url preview error getting response: %w", err)
	}

	return "", fmt.Errorf("url preview not supported for this url")
}
