package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

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

func unretweet(id uint64) error {
	req, err = http.NewRequest("POST", fmt.Sprintf("/1.1/statuses/unretweet/%d.json", id), nil)
	if err != nil {
		fmt.Println("Could not parse unretweet request: ", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Println("Could not send unretweet request: ", err)
		return err
	}

	b, err := ReadBody(resp)
	if err != nil {
		fmt.Println("UnRetweeting failed: ", err, string(b))
		return err
	}
	fmt.Println(string(b))

	return nil
}

func retweet(id uint64) error {
	req, err = http.NewRequest("POST", fmt.Sprintf("/1.1/statuses/retweet/%d.json", id), nil)
	if err != nil {
		fmt.Println("Could not parse retweet request: ", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Println("Could not send retweet request: ", err)
		return err
	}

	b, err := ReadBody(resp)
	if err != nil {
		fmt.Println("Retweeting failed: ", err, string(b))
		return err
	}

	return nil
}

func deleteTweet(id uint64) error {
	req, err = http.NewRequest("POST", fmt.Sprintf("/1.1/statuses/destroy/%d.json", id), nil)
	if err != nil {
		fmt.Println("Could not parse delete status request: ", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Println("Could not send delete status request: ", err)
		return err
	}

	b, err := ReadBody(resp)
	if err != nil {
		fmt.Println("Deleting status failed: ", err, string(b))
		return err
	}

	return nil
}

func searchReplies(fromId uint64) bool {
	query := url.Values{}
	query.Set("q", "to:"+user.ScreenName())
	query.Set("since_id", strconv.FormatUint(fromId, 10))
	query.Set("result_type", "recent")
	query.Set("count", "200")

	req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode()), nil)
	if err != nil {
		fmt.Println("Could not parse search request: ", err)
		return false
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Println("Could not send search request: ", err)
		return false
	}

	var searchResult struct {
		Statuses []twittergo.Tweet `json:"statuses"`
	}
	err = resp.Parse(&searchResult)
	if err != nil {
		fmt.Println("Problem parsing response: ", err)
		return false
	}

	// filter out tweets that are replies to other tweets
	fromIdStr := strconv.FormatUint(fromId, 10)
	for _, tweet := range searchResult.Statuses {
		if id, ok := tweet["in_reply_to_status_id_str"].(string); ok && id == fromIdStr {
			return true
		}
	}
	return false
}

func unretweetAll(all bool) error {
	var maxID uint64
	query := url.Values{}
	for {
		query.Set("screen_name", user.ScreenName())
		query.Set("user_id", fmt.Sprint(user.Id()))
		query.Set("count", "200")
		query.Set("tweet_mode", "extended")
		query.Set("include_entities", "true")
		if maxID > 0 {
			query.Set("max_id", fmt.Sprint(maxID))
		}
		req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/statuses/user_timeline.json?%v", query.Encode()), nil)
		if err != nil {
			fmt.Println("Could not parse user_timeline request: ", err)
			return err
		}

		resp, err = client.SendRequest(req)
		if err != nil {
			fmt.Println("Could not send user_timeline request: ", err)
			return err
		}

		// content, err := ioutil.ReadAll(resp.Body)
		// fmt.Println(string(content))

		var tweets []twittergo.Tweet
		err = resp.Parse(&tweets)
		if err != nil {
			fmt.Println("Problem parsing response: ", err)
			os.Exit(1)
		}

		fmt.Println("find tweets count:", len(tweets))
		for _, tweet := range tweets {
			//fmt.Println(tweet.Id(), "tweet", i, tweet.FullText())
			c, ok := tweet["created_at"]
			if !ok {
				//fmt.Println(tweet.Id(), "can't extract tweet created at")
				continue
			}
			createdAt, err := time.Parse(time.RubyDate, c.(string))
			if err != nil {
				//fmt.Println(tweet.Id(), "can't parse tweet created at")
				continue
			}
			if !createdAt.Add(24 * time.Hour).Before(time.Now()) {
				//fmt.Println(tweet.Id(), "tweet is posted in 24 hours", tweet.FullText())
				continue
			}

			if _, ok := tweet["in_reply_to_status_id_str"].(string); ok {
				//fmt.Println(tweet.Id(), "tweet is replied to other status", replyId, tweet.FullText())
				if rc, ok := tweet["retweet_count"].(uint64); ok && rc > 0 {
					//fmt.Println(tweet.Id(), "tweet is retweeted", tweet.FullText())
					continue
				}
				if fc, ok := tweet["favorite_count"].(uint64); ok && fc > 0 {
					//fmt.Println(tweet.Id(), "tweet is favorited", tweet.FullText())
					continue
				}
				if !searchReplies(tweet.Id()) {
					fmt.Println(tweet.Id(), "tweet is 0rt/0rp/0fav, should be deleted", tweet.FullText())
					// deleteTweet(tweet.Id())
				}
			}

			rs, ok := tweet["retweeted_status"]
			if !ok {
				//fmt.Println("can't find retweeted_status field")
				continue
			}
			fmt.Println(tweet.Id(), "find retweeted_status field")
			t, ok := rs.(map[string]interface{})
			if !ok {
				fmt.Println(tweet.Id(), "can't convert retweeted_status")
				continue
			}
			c, ok = t["created_at"]
			if !ok {
				fmt.Println(tweet.Id(), "can't extract retweeted_status created at")
				continue
			}
			createdAt, err = time.Parse(time.RubyDate, c.(string))
			if err != nil {
				fmt.Println(tweet.Id(), "can't parse retweeted_status created at")
				continue
			}
			if createdAt.Add(24 * time.Hour).Before(time.Now()) {
				idStr, ok := t["id_str"]
				if !ok {
					fmt.Println(tweet.Id(), "can't parse retweeted_status id str")
					continue
				}
				id, err := strconv.ParseUint(idStr.(string), 10, 64)
				if err != nil {
					fmt.Println(tweet.Id(), "can't convert retweeted_status id")
					continue
				}
				unretweet(id)
				fmt.Println(id, "is unretweeted")
				continue
			}
		}

		maxID = tweets[len(tweets)-1].Id()
		if !all {
			break
		}
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
	query.Set("count", "100")
	query.Set("tweet_mode", "extended")

	req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/statuses/user_timeline.json?%v", query.Encode()), nil)
	if err != nil {
		fmt.Println("Could not parse get user timeline request: ", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Println("Could not send get user timeline request: ", err)
		return err
	}

	var tweets []twittergo.Tweet
	err = resp.Parse(&tweets)
	if err != nil {
		fmt.Println("Problem parsing response:", err)
		os.Exit(1)
	}

	for _, tweet := range tweets {
		if tweet.CreatedAt().Add(24 * time.Hour).Before(time.Now()) {
			break
		}
		rt, ok := tweet["retweeted"]
		if !ok {
			fmt.Println("can't find retweet field")
			continue
		}
		if rt.(bool) {
			fmt.Println(tweet.Id(), "already retweeted")
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

		// check tag
		tags := tweet.Entities().Hashtags()
		for _, tag := range tags {
			text, ok := tag["text"]
			if ok {
				t, ok := text.(string)
				if ok && t == "FF" {
					// do retweet
					if err = retweet(tweet.Id()); err == nil {
						fmt.Println(tweet.Id(), "is retweeted")
						return nil
					}
				}
			}
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
			return nil
		}
	}

	return nil
}
