package main

import (
	"math/rand"
	"time"
)

type Skewed struct {
	hotUsers  []string
	coldUsers []string
}

func NewSkewed() *Skewed {
    return &Skewed{
        hotUsers: []string{
            "6b0d7d23-f4d0-4d2f-88d7-e3e7e31a4de8",
            "d1676227-ff16-448c-9e72-64ecdb8870fe",
            "4ea59087-d3a0-4a55-ba8b-c79c3a030b49",
        },
        coldUsers: []string{
            "589e0651-d05b-4325-a91d-e15276581b88",
            "718dd5bb-0536-459a-a4c3-5640efc6e564",
            "710b2dd5-9159-4c32-be20-3125cf51267e",
            "837a799b-f395-4a05-aedf-77393e203580",
            "48d3b680-3d2e-4682-b503-2a0bdc708248",
            "98530b7a-f7b1-4e65-b950-674187bf3084",
            "b4599173-008a-4d48-910f-419ff691638d",
            "83078c86-3c73-48df-a101-006ac1012c81",
        },
    }
}

func (s *Skewed) Next() RequestSpec {
	userID := ""
	if rand.Intn(100) < 75 {
		userID = s.hotUsers[rand.Intn(len(s.hotUsers))]
	} else {
		userID = s.coldUsers[rand.Intn(len(s.coldUsers))]
	}

	path := "/products/bbf855cb-6a60-4acb-829a-934a80c4b4b1"
	if rand.Intn(2) == 0 {
		path = "/orders/user/de46b506-2949-4b77-a27f-c250553b4263"
	}

	return RequestSpec{
		Method: "GET",
		Path:   path,
		UserID: userID,
	}
}

func (s *Skewed) Delay() time.Duration {
	return 12 * time.Millisecond
}