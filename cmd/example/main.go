package main

import (
	"fmt"
	"os"

	"github.com/OffBroadway/go-fatfs/pkg/fatfs"
)

func main() {
	img, err := fatfs.NewImageFile("/home/trevor/exfat-part.img")
	if err != nil {
		panic(err)
	}
	defer img.Close()

	fs, err := fatfs.NewFatFs(0)
	if err != nil {
		panic(err)
	}
	defer fs.Unmount()

	err = fs.Mount(img)
	if err != nil {
		panic(err)
	}

	dir, err := fs.Open("/")
	if err != nil {
		panic(err)
	}

	files, err := dir.Readdir(0)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		fmt.Println("FILE:", f.Name())
	}

	file, err := fs.OpenFile("/hello.txt", os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}

	data := []byte("Hello.... World?\n")
	_, err = file.Write(data)
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	file, err = fs.OpenFile("/hello.txt", os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	data = make([]byte, 512)
	n, err := file.Read(data)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}
