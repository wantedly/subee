package c

import (
	e "example.com/a/b/d"
)

type Book struct {
	Title  string   `json:"title"`
	Author e.Author `json:"author"`
}

