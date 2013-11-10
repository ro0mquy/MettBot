package plugins

import (
	"../ircclient"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	twitter_regex   = `\S*twitter\.com\/\S+\/status(es)?\/(\d+)\S*`
	twitter_api_url = `https://api.twitter.com/1.1/statuses/show.json?id=%v&trim_user=false&include_entities=true`
)

type tweet struct {
	Text string

	User struct {
		Screen_name string
	}

	Entities struct {
		Urls []struct {
			Url          string
			Expanded_url string
		}
	}

	Errors []struct {
		Message string
		Code    int
	}
}

type oauthAnswer struct {
	Token_type   string
	Access_token string
}

type TwitterPlugin struct {
	ic    *ircclient.IRCClient
	regex *regexp.Regexp
}

func (q *TwitterPlugin) String() string {
	return "twitter"
}

func (q *TwitterPlugin) Info() string {
	return "fetches content of posted tweets"
}

func (q *TwitterPlugin) Usage(cmd string) string {
	// just for interface saturation
	return ""
}

func (q *TwitterPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.regex = regexp.MustCompile(twitter_regex)
}

func (q *TwitterPlugin) Unregister() {
	return
}

func (q *TwitterPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "PRIVMSG" {
		//only process messages from chatrooms
		return
	}

	subs := q.regex.FindStringSubmatch(msg.Args[0])
	if subs == nil {
		// no url to tweet in message
		return
	}

	twt, err := q.fetchTweet(subs[2]) // second group of regex is searched tweet id
	if err != nil {
		log.Println(err)
		return
	}

	message := "@" + twt.User.Screen_name + ": " + twt.Text
	q.ic.ReplyMsg(msg, message)
}

func (q *TwitterPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (q *TwitterPlugin) fetchTweet(tweetId string) (twt tweet, err error) {
	tweetUrl := fmt.Sprintf(twitter_api_url, tweetId)
	oAuthToken := q.ic.GetStringOption("Twitter", "OAuthToken")
	if oAuthToken == "" {
		err = q.getOAuthToken()
		if err != nil {
			return
		}
		oAuthToken = q.ic.GetStringOption("Twitter", "OAuthToken")
	}

	request, err := http.NewRequest("GET", tweetUrl, nil)
	request.Header.Add("Authorization", "Bearer "+oAuthToken)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &twt)
	if err != nil {
		return
	}

	for _, e := range twt.Errors {
		if e.Code == 89 {
			err = fmt.Errorf("Twitter API returned: %s", e.Message)
			q.ic.RemoveOption("Twitter", "OAuthToken")
			return
		}
	}

	for _, u := range twt.Entities.Urls {
		twt.Text = strings.Replace(twt.Text, u.Url, u.Expanded_url, -1)
	}
	twt.Text = html.UnescapeString(twt.Text)

	if twt.User.Screen_name == "" || twt.Text == "" {
		err = fmt.Errorf("Error in tweet:\n%v\nUser: %v\nTweet text: %v", string(body), twt.User.Screen_name, twt.Text)
		return
	}

	return twt, nil
}

func (q *TwitterPlugin) getOAuthToken() (err error) {
	if q.ic.GetStringOption("Twitter", "OAuthToken") != "" {
		return nil
	}

	key := q.ic.GetStringOption("Twitter", "key")
	secret := q.ic.GetStringOption("Twitter", "secret")
	if key == "" || secret == "" {
		err = fmt.Errorf("To fetch tweets, you have to specify the key and a secret of your application")
		return
	}

	toBase64 := []byte(key + ":" + secret)
	encodedKey := base64.StdEncoding.EncodeToString(toBase64)
	bufferHTTPBody := bytes.NewBufferString("grant_type=client_credentials")

	requestOAuth, err := http.NewRequest("POST", "https://api.twitter.com/oauth2/token", bufferHTTPBody)
	requestOAuth.Header.Add("Authorization", "Basic "+encodedKey)
	requestOAuth.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	client := &http.Client{}
	responseOAuth, err := client.Do(requestOAuth)
	if err != nil {
		return
	}

	defer responseOAuth.Body.Close()
	bodyOAuth, err := ioutil.ReadAll(responseOAuth.Body)
	if err != nil {
		return
	}

	var answer oauthAnswer
	err = json.Unmarshal(bodyOAuth, &answer)
	if err != nil {
		return
	}
	if answer.Token_type != "bearer" {
		err = fmt.Errorf("False token type from Twitter OAuth API: %v", answer.Token_type)
		return
	}

	q.ic.SetStringOption("Twitter", "OAuthToken", answer.Access_token)
	return nil
}
