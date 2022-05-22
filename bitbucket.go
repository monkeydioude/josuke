package josuke

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Bitbucket push type and repository.
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
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
		} `json:"links"`
	} `json:"repository"`
}

func bitbucketToPayload(r io.Reader, hookEvent string) (*Payload, error) {
	var b Bitbucket
	err := json.NewDecoder(r).Decode(&b)

	if err != nil {
		return nil, err
	}

	// FIXME : remove this event name modification in future release,
	// made to avoid breaking change on 2022-05-20.
	if hookEvent == "repo:push" {
		hookEvent = "push"
	}

	var ref string
	if hookEvent == "push" {
		if len(b.Push.Changes) == 0 {
			return nil, errors.New("no push changes in payload for BitBucket push event")
		}
		ref = fmt.Sprintf("refs/heads/%s", b.Push.Changes[0].New.Name)
	} else {
		ref = ""
	}

	return &Payload{
		Ref:     ref,
		Action:  hookEvent,
		HtmlUrl: b.Repository.Links.HTML.Href,
		Repository: Repository{
			Name:    b.Repository.Fullname,
			HtmlUrl: b.Repository.Links.HTML.Href,
		},
	}, nil
}
