package model

type URL string
type Content []byte

type Result struct {
	URL     URL
	Content Content
}
