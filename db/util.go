package db

import (
	"regexp"
	"strings"
)

var removeSpace = regexp.MustCompile("\\s+")

func PrerareQuery(str string) string {
	str = strings.Replace(str, "\n", " ", -1)
	str = strings.Replace(str, "\t", " ", -1)
	str = removeSpace.ReplaceAllString(str, " ")
	return str
}
