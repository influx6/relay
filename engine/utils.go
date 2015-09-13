package engine

import (
	"regexp"
	"strconv"
	"time"
)

var elapso = regexp.MustCompile(`(\d+)(\w+)`)

func makeDuration(target string, def int) time.Duration {
	if !elapso.MatchString(target) {
		return time.Duration(def)
	}

	matchs := elapso.FindAllStringSubmatch(target, -1)

	if len(matchs) <= 0 {
		return time.Duration(def)
	}

	match := matchs[0]

	if len(match) < 3 {
		return time.Duration(def)
	}

	dur := time.Duration(ConvertToInt(match[1], def))

	mtype := match[2]

	switch mtype {
	case "s":
		return dur * time.Second
	case "mcs":
		return dur * time.Microsecond
	case "ns":
		return dur * time.Nanosecond
	case "ms":
		return dur * time.Millisecond
	case "m":
		return dur * time.Minute
	case "h":
		return dur * time.Hour
	default:
		return time.Duration(dur) * time.Second
	}

}

//ConvertToInt wraps the internal int coverter
func ConvertToInt(target string, def int) int {
	fo, err := strconv.Atoi(target)
	if err != nil {
		return def
	}
	return fo
}
