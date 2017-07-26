package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GetFriendsList() ([]UserProfile, error) {
	var cursor int64 = -1
	var users []UserProfile
	for {
		query := url.Values{}
		query.Set("screen_name", user.ScreenName())
		query.Set("user_id", fmt.Sprintf("%d", user.Id()))
		query.Set("cursor", fmt.Sprintf("%d", cursor))
		query.Set("count", "2000")
		req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/friends/list.json?%v", query.Encode()), nil)
		if err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			return nil, err
		}

		resp, err = client.SendRequest(req)
		if err != nil {
			fmt.Printf("Could not send request: %v\n", err)
			return nil, err
		}
		userColl := new(UserCollection)

		if b, err := ReadBody(resp); err != nil {
			return nil, err
		} else {
			err = json.Unmarshal(b, userColl)
			if err == io.EOF {
				err = nil
			}
			if err != nil {
				fmt.Printf("Problem parsing friends response: %v\n", err)
				return nil, err
			}
		}

		if len(userColl.Users) < 1 {
			break
		}
		users = append(users, userColl.Users...)
		cursor = userColl.NextCursor
	}
	return users, nil
}

func GetFollowersList() ([]UserProfile, error) {
	/// get followers list
	var cursor int64 = -1
	var users []UserProfile
	for {
		query := url.Values{}
		query.Set("screen_name", user.ScreenName())
		query.Set("user_id", fmt.Sprintf("%d", user.Id()))
		query.Set("cursor", fmt.Sprintf("%d", cursor))
		query.Set("count", "2000")
		req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/followers/list.json?%v", query.Encode()), nil)
		if err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			return nil, err
		}

		resp, err = client.SendRequest(req)
		if err != nil {
			fmt.Printf("Could not send request: %v\n", err)
			return nil, err
		}
		userColl := new(UserCollection)

		if b, err := ReadBody(resp); err != nil {
			return nil, err
		} else {
			err = json.Unmarshal(b, userColl)
			if err == io.EOF {
				err = nil
			}
			if err != nil {
				fmt.Printf("Problem parsing followers response: %v\n", err)
				//fmt.Println(string(b))
				return nil, err
			}
		}

		if len(userColl.Users) < 1 {
			break
		}
		users = append(users, userColl.Users...)
		cursor = userColl.NextCursor
	}
	return users, nil
}
