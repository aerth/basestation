package basestation

import (
	"fmt"
	"io"
	"log"
)

func Test() {
	s := "testing"

	log.Println("Boot #", hitCounterRead(s))
}

type File struct {
	objectid string
	data     []byte
	io.Writer
	io.Reader
}

// Find the index of string or return 404 error
func Index(a []string, s string) (int, error) {
	for i, k := range a {
		if k == s {
			return i, nil
		}
	}
	return len(a), fmt.Errorf("404")
}

// Remove removes from slice if staisfy f()
// (Example: only() or everythingbut())
// Strings.
func Remove(s []string, fn func(string, string) bool, g string) []string {
	var p []string // == nil
	for _, v := range s {
		if fn(g, v) {
			p = append(p, v)
		}
	}
	return p
}

func everythingbut(s, g string) bool {
	return s == g
}
func only(s, g string) bool {
	return s != g
}

// Filter returns a new slice holding only
// the elements of s that satisfy f()
// int
func Filter(s []int, fn func(int) bool) []int {
	var p []int // == nil
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

func evens(i int) bool {
	if i%2 != 0 {
		return false
	}
	return true
}
func odds(i int) bool {
	if i%2 != 1 {
		return false
	}
	return true
}
