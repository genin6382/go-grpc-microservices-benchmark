package main

import (
	"math/rand"
	"time"
)

type Bursty struct {
	paths []string
	count int
}

func NewBursty() *Bursty {
	return &Bursty{
		paths: []string{
			"/users/login",
			"/products/de46b506-2949-4b77-a27f-c250553b4263",
			"/orders/57ee156e-8628-429a-803e-bec229d1e244",
		},
	}
}

func (b *Bursty) Next() RequestSpec {
	b.count++
	p := b.paths[rand.Intn(len(b.paths))]
	spec := RequestSpec{
		Method: "GET",
		Path:   p,
	}

	if p == "/users/login" {
		spec.Method = "POST"
		spec.Body = `{"name":"Vidhu","password":"Vidhu$12"}`
	} else {
		spec.UserID = "bursty-user"
	}

	return spec
}

func (b *Bursty) Delay() time.Duration {
	if b.count%30 < 8 {
		return 2 * time.Millisecond
	}
	return 90 * time.Millisecond
}