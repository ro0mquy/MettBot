package answers

import "math/rand"

func RandStr(str string) string {
	section := answers[str]
	index := rand.Intn(len(section))
	return section[index]
}

var answers = map[string][]string{
	"dong": {
		"DONG",
		"DÖNG",
		"DÄNG",
		"DING",
		"PLÖNK",
		"BÄM",
		"KLONK",
	},
	"addedQuote": { // needs %v which is a number
		"Added quote #%v to database",
		"The %vth quote which will be never forgotten",
		"Another quote from the endless depths of #mett (#%v)",
		"And we've got quote number %v",
	},
	"addedMett": { // needs %v which is a number
		"Added mett #%v to database",
		"Now I've already %v entries of mettcontent",
	},
}
