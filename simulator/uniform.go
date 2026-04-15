package main

import (
	"math/rand"
	"time"
)

type Uniform struct {
	paths []string
}

func NewUniform() *Uniform {
	return &Uniform{
		paths: []string{
			"/users/login",
			"/products/de46b506-2949-4b77-a27f-c250553b4263",
			"/orders/57ee156e-8628-429a-803e-bec229d1e244",
			"/orders/user/de46b506-2949-4b77-a27f-c250553b4263",
		},
	}
}

func (u *Uniform) Next() RequestSpec {
	p := u.paths[rand.Intn(len(u.paths))]
	spec := RequestSpec{
		Method: "GET",
		Path:   p,
	}

	if p == "/users/login" {
		spec.Method = "POST"
		spec.Body = `{"name":"Vidhu","password":"Vidhu$12"}`
	} else {
		spec.UserID = "uniform-user"
	}

	return spec
}

func (u *Uniform) Delay() time.Duration {
	return 25 * time.Millisecond
}