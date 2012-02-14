package main

// Zum Rumspielen mit XML-Parsern. Einfacher, als das in nem Plugin zu machen

import (
	"http"
	"fmt"
	"xml"
)

type root struct {
	Events Events
}

type Events struct {
	Event []Event
}

type Event struct {
	Id         int `xml:"attr"`
	Submission Submission
	Judging    Judging
}

type Submission struct {
	Id       int `xml:"attr"`
	Team     string
	Problem  string
	Language string
}

type Judging struct {
	Id       int    `xml:"attr"`
	Submitid int    `xml:"attr"`
	Result   string `xml:"chardata"`
}

func main() {
	var client http.Client
	//response, _ := client.Get("https://bot:hallowelt@icpc.informatik.uni-erlangen.de/domjudge/plugin/event.php")
	// Um den DOMJudge nicht uebermaessig in der Entwicklungsphase zu pollen 
	// TODO: Konfigurierbar
	response, _ := client.Get("http://d-paulus.de/tmp.xml")
	fmt.Println(response.StatusCode)
	var res root
	xml.Unmarshal(response.Body, &res)
	response.Body.Close()
	fmt.Println(len(res.Events.Event))
	//fmt.Println(res.Events.Event[482].Judging.Result)
	//for i := 0; i < len(res.Events.Event); i = i + 1 {
	//	fmt.Println(res.Events.Event[i].Id)
	//}
	//fmt.Println(content)
}
