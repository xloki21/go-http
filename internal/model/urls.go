package model

// todo: переделать с помощью stdlib url.URL?

type URL string
type Content []byte

type Result struct {
	URL     URL
	Content Content
}
