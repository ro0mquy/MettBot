package answers

import "math/rand"

func RandStr(str string) string {
	section := answers[str]
	index := rand.Intn(len(section))
	return section[index]
}

var answers = map[string][]string {
	"dong": {
		"DONG",
		"DÖNG",
		"DÄNG",
		"DING",
		"PLÖNK",
		"BÄM",
		"KLONK",
	},
}
