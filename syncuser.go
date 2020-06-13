package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/kurrik/twittergo"
)

func findUser() (uint64, error) {
	query := url.Values{}
	query.Set("screen_name", "freshfruitcn")

	req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/users/lookup.json?%v", query.Encode()), nil)
	if err != nil {
		fmt.Printf("Could not parse sync user request: %v\n", err)
		return 0, err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send sync user request: %v\n", err)
		return 0, err
	}

	var users []twittergo.User
	err = resp.Parse(&users)
	if err != nil {
		fmt.Printf("Problem parsing response: %v\n", err)
		os.Exit(1)
	}
	if len(users) != 1 {
		fmt.Println("Can't find the user @freshfruitcn")
		os.Exit(1)
	}
	user := users[0]
	fmt.Printf("ID: %v\n", user.Id())
	fmt.Printf("Name: %v\n", user.Name())
	fmt.Printf("ScreenName: %v\n", user.ScreenName())
	return user.Id(), nil
}

func retweet(id uint64) error {
	req, err = http.NewRequest("POST", fmt.Sprintf("/1.1/statuses/retweet/%d.json", id), nil)
	if err != nil {
		fmt.Printf("Could not parse sync user request: %v\n", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send sync user request: %v\n", err)
		return err
	}

	b, err := ReadBody(resp)
	if err != nil {
		fmt.Printf("Retweeting failed: %v\n", err, string(b))
		return err
	}

	return nil
}

func syncUser() error {
	userID, err := findUser()
	if err != nil {
		os.Exit(1)
	}
	query := url.Values{}
	query.Set("screen_name", "freshfruitcn")
	query.Set("user_id", fmt.Sprint(userID))
	query.Set("count", "10")

	req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/statuses/user_timeline.json?%v", query.Encode()), nil)
	if err != nil {
		fmt.Printf("Could not parse get user timeline request: %v\n", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send get user timeline request: %v\n", err)
		return err
	}

	var tweets []twittergo.Tweet
	err = resp.Parse(&tweets)
	if err != nil {
		fmt.Printf("Problem parsing response: %v\n", err)
		os.Exit(1)
	}

	for _, tweet := range tweets {
		rt, ok := tweet["retweeted"]
		if !ok {
			fmt.Println("can't find retweet field")
			continue
		}
		if rt.(bool) == true {
			fmt.Println("already retweeted")
			continue
		}
		// check user
		user, ok := tweet["user"]
		if !ok {
			fmt.Println("can't find user field")
			continue
		}
		u, ok := user.(map[string]interface{})
		if !ok {
			fmt.Println("converting user struct failed", user)
			continue
		}
		id, ok := u["id"]
		if !ok {
			fmt.Println("can't find user id field")
			continue
		}
		uid, ok := id.(float64)
		if !ok {
			fmt.Println("converting user id type failed", id)
			continue
		}
		if uint64(uid) != userID {
			fmt.Println("not posted by user", userID)
			continue
		}
		// check has media
		ee := tweet.ExtendedEntities()
		if len(ee) == 0 {
			fmt.Println("no extended entities, skip")
			continue
		}

		// do retweet
		if err = retweet(tweet.Id()); err == nil {
			fmt.Println(tweet.Id(), "is retweeted")
		}
	}

	return nil
}
