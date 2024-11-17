package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/disintegration/imaging"
)

func main() {
	src := os.Args[1]
	srcImg, err := imaging.Open(src)

	if err != nil {
		panic("invalid image")
	}

	img := imaging.Resize(srcImg, 250, 500, imaging.Lanczos)
	filename := strings.TrimSuffix(src, path.Ext(src))
	newFilename := fmt.Sprintf("%s_%dx%d.png", filename, 250, 500)

	err = imaging.Save(img, newFilename)
	if err != nil {
		panic("error saving image")
	}

	return
}
