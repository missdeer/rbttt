package main

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

func Download(u string) error {
	return nil
}

func UpdateProfileBackgroundImage() {
	// first download all friends profile images
	users, err := GetFollowersList()
	if err != nil {
		os.Exit(1)
	}

	for _, v := range users.Users {
		Download(v.ProfileImageUrl) // download
	}

	// input files
	files := []string{"01.jpg", "02.jpg", "03.jpg"}

	// load images and make 100x100 thumbnails of them
	var thumbnails []image.Image
	for _, file := range files {
		img, err := imaging.Open(file)
		if err != nil {
			panic(err)
		}
		thumb := imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
		thumbnails = append(thumbnails, thumb)
	}
	// create a new blank image
	dst := imaging.New(100*len(thumbnails), 100, color.NRGBA{0, 0, 0, 0})

	// paste thumbnails into the new image side by side
	for i, thumb := range thumbnails {
		dst = imaging.Paste(dst, thumb, image.Pt(i*100, 0))
	}

	// save the combined image to file

	if err := imaging.Save(dst, "dst.jpg"); err != nil {
		panic(err)
	}
}
