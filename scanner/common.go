package scanner

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger(verbose bool) {
	if verbose {
		Logger.SetLevel(logrus.DebugLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
	Logger.SetOutput(os.Stdout)
}

func removeURLsFromContent(content []byte, urls []string) []byte {
	modified := string(content)
	for _, u := range urls {
		modified = ReplaceAllSafe(modified, u, "")
	}
	return []byte(modified)
}

func ReplaceAllSafe(s, old, new string) string {
	if old == "" {
		return s
	}
	return strings.ReplaceAll(s, old, new)
}

func replaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}
