package answers

import "math/rand"

func RandStr(str []string) string {
	index := rand.Intn(len(str))
	return str[index]
}

var Warning []string = []string{
	"I warn you",
	"Back to Mett",
}

var Syntax []string = []string{
	"Wrong Syntax. Try !help",
}

var AddedQuote []string = []string{	// needs %v which is a number
	"Added quote #%v to database",
}

var AddedMett []string = []string{	// needs %v which is a number
	"Added mett #%v to database",
}

var OffendNick []string = []string{	// (ACTION) needs %v which is a string
	"slaps %v",
}

var QuoteNotFound []string = []string{
	"Quote not found",
}

var MettTime []string = []string{	// needs %v which is a string
	"It's time for moar mett: %v",
}

var IgnoreCmd []string = []string{
	"Bother someone else!",
	"You're annoying me",
	"No way",
}
