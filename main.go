package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	flag "github.com/ogier/pflag"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	err    error
	client *twittergo.Client
	req    *http.Request
	resp   *twittergo.APIResponse
	user   *twittergo.User
)

type UserProfile struct {
	Id                  int64  `json:"id"`
	Name                string `json:"name"`
	ScreenName          string `json:"screen_name"`
	ProfileImageUrl     string `json:"profile_image_url"`
	DefaultProfileImage bool   `json:"default_profile_image"`
	DefaultProfile      bool   `json:"default_profile"`
	StatusesCount       int    `json:"statuses_count"`
}

type UserCollection struct {
	Users             []UserProfile `json:"users"`
	NextCursor        int64         `json:"next_cursor"`
	NextCursorStr     string        `json:"next_cursor_str"`
	PreviousCursor    int64         `json:"previous_cursor"`
	PreviousCursorStr string        `json:"previous_cursor_str"`
}

func ReadBody(r *twittergo.APIResponse) (b []byte, err error) {
	var (
		header string
		reader io.Reader
	)
	defer r.Body.Close()
	header = strings.ToLower(r.Header.Get("Content-Encoding"))
	if header == "" || strings.Index(header, "gzip") == -1 {
		reader = r.Body
	} else {
		if reader, err = gzip.NewReader(r.Body); err != nil {
			return
		}
	}
	b, err = ioutil.ReadAll(reader)
	return
}

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

func Authorize() {
	config := &oauth1a.ClientConfig{
		ConsumerKey:    "5AD7pjFonMS6JIevrBz1Q",
		ConsumerSecret: "wKdiQ2kPZMo2Q1uK71qv4KkW7L8NyLkbubTfh87ZU",
		CallbackURL:    "oob",
	}
	// read access token key & access token secret from file
	credentials, err := ioutil.ReadFile(".CREDENTIALS")
	if err == nil {
		lines := strings.Split(string(credentials), "\n")

		// load access token key & access token secret
		auth := oauth1a.NewAuthorizedConfig(lines[0], lines[1])
		client = twittergo.NewClient(config, auth)
	} else {
		service := &oauth1a.Service{
			RequestURL:   "https://api.twitter.com/oauth/request_token",
			AuthorizeURL: "https://api.twitter.com/oauth/authorize",
			AccessURL:    "https://api.twitter.com/oauth/access_token",
			ClientConfig: config,
			Signer:       new(oauth1a.HmacSha1Signer),
		}

		httpClient := new(http.Client)
		userConfig := &oauth1a.UserConfig{}
		userConfig.GetRequestToken(service, httpClient)
		u, _ := userConfig.GetAuthorizeURL(service)
		fmt.Println("use a web browser to open", u)
		token, _ := userConfig.GetToken()
		var verifier string
		fmt.Printf("input PIN code: ")
		fmt.Scanf("%s", &verifier)
		if err := userConfig.GetAccessToken(token, verifier, service, httpClient); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// save access token key & access token secret to file
		f, err := os.OpenFile(".CREDENTIALS", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString(userConfig.AccessTokenKey)
			f.WriteString("\n")
			f.WriteString(userConfig.AccessTokenSecret)
			f.Close()
			fmt.Println("save auth info into .CREDENTIALS")
		}

		// load access token key & access token secret
		auth := oauth1a.NewAuthorizedConfig(userConfig.AccessTokenKey, userConfig.AccessTokenSecret)
		client = twittergo.NewClient(config, auth)
	}

	req, err = http.NewRequest("GET", "/1.1/account/verify_credentials.json", nil)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}

	user = &twittergo.User{}
	err = resp.Parse(user)
	if err != nil {
		fmt.Printf("Problem parsing response: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("ID: %v\n", user.Id())
	fmt.Printf("Name: %v\n", user.Name())
	fmt.Printf("ScreenName: %v\n", user.ScreenName())
}

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
			} else {
				time.Sleep(10 * time.Second)
				goto try_unblock
			}
			time.Sleep(5 * time.Second)
		}
		i += len(users.Users)
		if len(users.Users) < 1 {
			break
		}
		cursor = users.NextCursor
		time.Sleep(30 * time.Second)
	}
	fmt.Printf("unblocked %d users\n", i)
}

func init() {
}

func main() {
	fmt.Println("rbttt, the small twitter helper tool.")

	unblock := false
	block := false
	background := false
	all := false

	flag.BoolVarP(&unblock, "unblock", "u", false, "clear block list")
	flag.BoolVarP(&block, "block", "b", false, "block followers who are using default profile image or have 0 tweet so far")
	flag.BoolVarP(&background, "backgroud", "g", false, "update profile background image with friends' avantar wall")
	flag.BoolVarP(&all, "all", "a", false, "run all actions")

	flag.Parse()

	Authorize()

	if all == true || block == true {
		BlockUnexpectedUsers()
	}

	if all == true || unblock == true {
		ClearBlockList()
	}

	if all == true || background == true {
		UpdateProfileBackgroundImage()
	}
}
