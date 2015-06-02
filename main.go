package main

import (
	"compress/gzip"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	flag "github.com/ogier/pflag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	err             error
	client          *twittergo.Client
	req             *http.Request
	resp            *twittergo.APIResponse
	user            *twittergo.User
	credentialsFile string
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

func Authorize(force_auth bool) {
	config := &oauth1a.ClientConfig{
		ConsumerKey:    "5AD7pjFonMS6JIevrBz1Q",
		ConsumerSecret: "wKdiQ2kPZMo2Q1uK71qv4KkW7L8NyLkbubTfh87ZU",
		CallbackURL:    "oob",
	}
	// read access token key & access token secret from file
	credentials, err := ioutil.ReadFile(credentialsFile)
	if err == nil && force_auth == false {
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
		f, err := os.OpenFile(credentialsFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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

func main() {
	fmt.Println("rbttt, the small twitter helper tool.")

	unblock := false
	block := false
	background := false
	all := false
	reauth := false

	flag.BoolVarP(&unblock, "unblock", "u", false, "clear block list")
	flag.BoolVarP(&block, "block", "b", false, "block followers who are using default profile image or have 0 tweet so far")
	flag.BoolVarP(&background, "backgroud", "g", false, "update profile background image with friends' avantar wall")
	flag.BoolVarP(&all, "all", "a", false, "run all actions")
	flag.BoolVarP(&reauth, "reauth", "r", false, "re-authenticate current credential")
	flag.StringVarP(&credentialsFile, "config", "c", ".CREDENTIALS", "set configuration file which contains credentials")

	flag.Parse()

	Authorize(reauth)

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
