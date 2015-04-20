package main

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/oauth2"
	"os"
)

func main() {
	// create an http.Client which handles authentication

	// for Twitter "app-only auth" use golang/oauth2 http client
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	if accessToken == "" {
		panic("Missing TWITTER_ACCESS_TOKEN environment variable")
	}
	ts := &tokenSource{&oauth2.Token{AccessToken: accessToken}}
	appAuthClient := oauth2.NewClient(oauth2.NoContext, ts)

	// Twitter

	client := twitter.NewClient(appAuthClient)

	// user show
	userShowParams := &twitter.UserShowParams{ScreenName: "golang"}
	user, _, _ := client.Users.Show(userShowParams)
	fmt.Printf("users/show:\n%+v\n", user)

	// users lookup
	userLookupParams := &twitter.UserLookupParams{ScreenName: []string{"golang", "gophercon"}}
	users, _, _ := client.Users.Lookup(userLookupParams)
	fmt.Printf("users/lookup:\n%+v\n", users)

	// status show
	statusShowParams := &twitter.StatusShowParams{}
	tweet, _, _ := client.Statuses.Show(584077528026849280, statusShowParams)
	fmt.Printf("statuses/show:\n%+v\n", tweet)

	// statuses lookup
	statusLookupParams := &twitter.StatusLookupParams{Id: []int64{20}}
	tweets, _, _ := client.Statuses.Lookup([]int64{573893817000140800}, statusLookupParams)
	fmt.Printf("statuses/lookup:\n%v\n", tweets)

	// user timeline
	userTimelineParams := &twitter.UserTimelineParams{ScreenName: "golang", Count: 2}
	tweets, _, _ = client.Timelines.UserTimeline(userTimelineParams)
	fmt.Printf("statuses/user_timeline:\n%+v\n", tweets)
}

// golang/oauth2

type tokenSource struct {
	token *oauth2.Token
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}