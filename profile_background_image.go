package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/kurrik/twittergo"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	profileBackgroundImage = "profileBackgroundImage.jpg"
)

type Media struct {
	MediaId    int64  `json:"media_id"`
	MediaIdStr string `json:"media_id_str"`
}

func Download(requestUrl string) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		fmt.Println("make request failed", err)
		return []byte{}, err
	}

	//	if len(referer) > 0 {
	//		req.Header.Set("Referer", referer)
	//	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:34.0) Gecko/20100101 Firefox/34.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("getting data from %s failed %s", requestUrl, err.Error())
		return []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("reading data failed ", err)
		return []byte{}, err
	}

	return body, nil
}

func GetBody() (body io.ReadWriter, header string, err error) {
	var (
		mp     *multipart.Writer
		media  []byte
		writer io.Writer
	)
	body = bytes.NewBufferString("")
	mp = multipart.NewWriter(body)
	media, err = ioutil.ReadFile(profileBackgroundImage)
	if err != nil {
		return
	}

	mp.WriteField("media_data", base64.StdEncoding.EncodeToString(media))

	writer, err = mp.CreateFormFile("media[]", profileBackgroundImage)
	if err != nil {
		return
	}
	writer.Write(media)
	header = fmt.Sprintf("multipart/form-data;boundary=%v", mp.Boundary())
	mp.Close()
	return
}

func GenerateProfileBackgroundImage() {
	// first download all friends profile images
	users, err := GetFriendsList()
	if err != nil {
		os.Exit(1)
	}

	const width = 42 - 10
	const height = 30 - 10
	// create a new blank image
	dst := imaging.New(width*48, height*48, color.NRGBA{0, 0, 0, 0})
	// input files
	var x int = 0
	var y int = 0
gen_background_image:
	for _, v := range users {
		if b, err := Download(v.ProfileImageUrl); err == nil {
			r := bytes.NewReader(b)
			if img, _, err := image.Decode(r); err == nil {
				fmt.Println("paste ", v.ProfileImageUrl)
				dst = imaging.Paste(dst, img, image.Pt(x*48, y*48))
				if x++; x >= width {
					x = 0
					y++
				}
				if y >= height {
					break
				}
			}
		}
	}
	if x < width-1 || y < height-1 {
		goto gen_background_image
	}

	// save the combined image to file
	if err := imaging.Save(dst, profileBackgroundImage); err != nil {
		fmt.Println("saving the final image file failed", err)
		os.Exit(1)
	}
}

func UpdateProfileBackgroundImage() {
	GenerateProfileBackgroundImage()

	// upload to twitter
	body, header, err := GetBody()
	if err != nil {
		fmt.Printf("Problem loading body: %v\n", err)
		os.Exit(1)
	}

	req, err = http.NewRequest("POST", "https://upload.twitter.com/1.1/media/upload.json", body)
	if err != nil {
		fmt.Printf("Could not parse uploading media request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", header)

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send uploading media request: %v\n", err)
		os.Exit(1)
	}

	media := new(Media)
	b, err := ReadBody(resp)
	if err != nil {
		fmt.Println("reading body failed", err)
		os.Exit(1)
	}

	err = json.Unmarshal(b, media)
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		fmt.Printf("Problem parsing media response: %v\n", err)
		os.Exit(1)
	}

	requestUrl := fmt.Sprintf(`/1.1/account/update_profile_background_image.json?skip_status=1&tile=1&media_id=%d`, media.MediaId)
	req, err = http.NewRequest("POST", requestUrl, nil)
	if err != nil {
		fmt.Printf("Could not parse updating profile background image request: %v\n", err)
		os.Exit(1)
	}

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send updating profile background image request: %v\n", err)
		os.Exit(1)
	} else {
		if resp.StatusCode == twittergo.STATUS_OK {
			fmt.Println("profile background image updated")
		} else {
			fmt.Println(resp)
		}
	}

}
