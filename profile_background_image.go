package main

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

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

func UpdateProfileBackgroundImage() {
	// first download all friends profile images
	users, err := GetFriendsList()
	if err != nil {
		os.Exit(1)
	}

	const width = 42
	var height = len(users) / width
	if height < 20 {
		height = 20
	}
	if height > 30 {
		height = 30
	}
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
	if err := imaging.Save(dst, "dst.png"); err != nil {
		fmt.Println("saving the final image file failed", err)
	} else {
		fmt.Println("got dst.png")
	}
}
