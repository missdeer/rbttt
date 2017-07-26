package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func BlockUser(id int64, screen_name string) error {
	query := url.Values{}
	query.Set("screen_name", screen_name)
	query.Set("user_id", fmt.Sprintf("%d", id))
	query.Set("skip_status ", "true")

	req, err = http.NewRequest("POST", fmt.Sprintf("/1.1/blocks/create.json?%v", query.Encode()), nil)
	if err != nil {
		fmt.Printf("Could not parse block request: %v\n", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send block request: %v\n", err)
		return err
	}

	return nil
}

func UnblockUser(id int64, screen_name string) error {
	query := url.Values{}
	query.Set("screen_name", screen_name)
	query.Set("user_id", fmt.Sprintf("%d", id))
	query.Set("skip_status ", "true")

	req, err = http.NewRequest("POST", fmt.Sprintf("/1.1/blocks/destroy.json?%v", query.Encode()), nil)
	if err != nil {
		fmt.Printf("Could not parse unblock request: %v\n", err)
		return err
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send unblock request: %v\n", err)
		return err
	}

	return nil
}

func BlockUnfollowingUsers() {
	followers, err := GetFollowersList()
	if err != nil {
		os.Exit(1)
	}
	friends, err := GetFriendsList()
	if err != nil {
		os.Exit(1)
	}
	i := 0
	for _, follower := range followers {
		isFriend := false
		for _, friend := range friends {
			if follower.Id == friend.Id {
				isFriend = true
				break
			}
		}

		if !isFriend {
			i++
		try_block:
			if err = BlockUser(follower.Id, follower.ScreenName); err == nil {
				fmt.Printf("id: %v, screen name: %s, name: %s, profile image url: %s, default image: %v, default profile: %v, statuses count: %d has been blocked because I'm not following\n",
					follower.Id, follower.ScreenName, follower.Name, follower.ProfileImageUrl, follower.DefaultProfileImage, follower.DefaultProfile, follower.StatusesCount)
			} else {
				time.Sleep(10 * time.Second)
				goto try_block
			}
		}
	}
	fmt.Printf("blocked %d followers whom I'm not following\n", i)
}

func BlockUnexpectedUsers() {
	users, err := GetFollowersList()
	if err != nil {
		os.Exit(1)
	}
	var i int = 0
	for _, v := range users {
		if v.DefaultProfileImage == true || v.StatusesCount == 0 {
			i++
		try_block:
			if err = BlockUser(v.Id, v.ScreenName); err == nil {
				fmt.Printf("id: %v, screen name: %s, name: %s, profile image url: %s, default image: %v, default profile: %v, statuses count: %d has been blocked\n",
					v.Id, v.ScreenName, v.Name, v.ProfileImageUrl, v.DefaultProfileImage, v.DefaultProfile, v.StatusesCount)
			} else {
				time.Sleep(10 * time.Second)
				goto try_block
			}
		}
	}
	fmt.Printf("blocked %d followers who are using default profile image or have 0 tweet posted\n", i)
}

func ClearBlockList() {
	/// get block list
	var cursor int64 = -1
	var i int = 0
	for {
		query := url.Values{}
		query.Set("cursor", fmt.Sprintf("%d", cursor))
		query.Set("skip_status", "true")
		req, err = http.NewRequest("GET", fmt.Sprintf("/1.1/blocks/list.json?%v", query.Encode()), nil)
		if err != nil {
			fmt.Printf("Could not parse block list request: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		resp, err = client.SendRequest(req)
		if err != nil {
			fmt.Printf("Could not send block list request: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		users := new(UserCollection)

		if b, err := ReadBody(resp); err != nil {
			time.Sleep(5 * time.Second)
			continue
		} else {
			err = json.Unmarshal(b, users)
			if err == io.EOF {
				err = nil
			}
			if err != nil {
				fmt.Printf("Problem parsing block list response: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			}
		}

		for _, v := range users.Users {
		try_unblock:
			if err = UnblockUser(v.Id, v.ScreenName); err == nil {
				fmt.Printf("id: %v, screen name: %s, name: %s, profile image url: %s, default image: %v, default profile: %v has been unblocked\n",
					v.Id, v.ScreenName, v.Name, v.ProfileImageUrl, v.DefaultProfileImage, v.DefaultProfile)
				time.Sleep(5 * time.Second)
			} else {
				time.Sleep(10 * time.Second)
				goto try_unblock
			}
		}
		i += len(users.Users)
		if len(users.Users) < 1 {
			break
		}
		cursor = users.NextCursor
	}
	fmt.Printf("unblocked %d users\n", i)
}
