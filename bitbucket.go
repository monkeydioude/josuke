package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type Bitbucket struct {
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
}

func bitbucketToPayload(r io.Reader) *Payload {
	var b Bitbucket
	err := json.NewDecoder(r).Decode(&b)

	if err != nil {
		log.Printf("[ERR ] %s", err)
		return nil
	}

	if len(b.Push.Changes) == 0 {
		log.Println("[ERR ] Could not decode body into Payload")
		return nil
	}

	return &Payload{
		Ref:     fmt.Sprintf("refs/heads/%s", b.Push.Changes[0].New.Name),
		Action:  "push",
		HtmlUrl: b.Repository.Links.Html.Href,
		Repository: Repository{
			Name:    b.Repository.Fullname,
			HtmlUrl: b.Repository.Links.Html.Href,
		},
	}
}
