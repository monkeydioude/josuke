package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type Bitbucket struct {
	Data struct {
		Push struct {
			Changes []struct {
				New struct {
					Type string `json:"type"`
					Name string `json:"name"`
				} `json:"new"`
			} `json:"changes"`
		} `json:"push"`
		Repository struct {
			Fullname string `json:"full_name"`
			Links    struct {
				Html struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"links"`
		} `json:"repository"`
	} `json:"data"`
}

func bitbucketToPayload(r io.Reader) *Payload {
	var b Bitbucket
	err := json.NewDecoder(r).Decode(&b)

	if err != nil {
		log.Printf("[WARN] %s", err)
		return nil
	}

	return &Payload{
		Ref:     fmt.Sprintf("refs/heads/%s", b.Data.Push.Changes[0].New.Name),
		Action:  "push",
		HtmlUrl: b.Data.Repository.Links.Html.Href,
		Repository: Repository{
			Name:    b.Data.Repository.Fullname,
			HtmlUrl: b.Data.Repository.Links.Html.Href,
		},
	}
}
