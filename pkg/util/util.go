package util

import (
	"errors"
	"strings"

	"github.com/valyala/fasthttp"
)

const (
	CHANNELS   int = 2
	FRAME_RATE int = 48000
	FRAME_SIZE int = 960
	MAX_BYTES  int = (FRAME_SIZE * 2) * 2

	ARG_IDENTIFIER string = " "
	COMPLETE       string = "✅"
)

var (
	httpClient fasthttp.Client = fasthttp.Client{}
)

func AfterCommand(message string) string {
	i := strings.Index(message, ARG_IDENTIFIER)
	if i == -1 || i+1 >= len(message) {
		return ""
	}
	return message[i+1:]
}

func DoWithRedirects(req *fasthttp.Request, res *fasthttp.Response) error {
	for i := 0; i <= 5; i++ {
		err := httpClient.Do(req, res)
		if err != nil {
			return err
		}
		if !fasthttp.StatusCodeIsRedirect(res.StatusCode()) {
			return nil
		}
		location := res.Header.Peek("Location")
		if len(location) == 0 {
			return errors.New("missing 'Location' header after redirect")
		}
		req.URI().UpdateBytes(location)
		res.Reset()
	}
	return errors.New("redirected too many times")
}

// FindJson parses a string and returns the first instance
// of a json object indicated by the first { and last }
func FindJson(s string) string {
	bracketStack := []rune{}
	var start, end int

	for i, c := range s {
		if c == '{' {
			if len(bracketStack) == 0 {
				start = i
			}
			bracketStack = append(bracketStack, '{')
		}
		if c == '}' {
			bracketStack = bracketStack[:len(bracketStack)-1]
			if len(bracketStack) == 0 {
				end = i
				break
			}
		}
	}
	return s[start : end+1]
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
