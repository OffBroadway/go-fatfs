package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/OffBroadway/go-fatfs/pkg/fatfs"
	"github.com/spf13/afero"

	ftpserver "github.com/fclairamb/ftpserverlib"
	logrus "github.com/fclairamb/go-log/logrus"
)

func main() {
	img, err := fatfs.NewImageFile("/home/trevor/exfat-part.img")
	if err != nil {
		panic(err)
	}

	fatfs.RegisterBlockDevice(0, img)

	fs, err := fatfs.NewFatFs()
	if err != nil {
		panic(err)
	}

	err = fs.Mount("0:", 1)
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

	fmt.Println("DATA:", string(data[:n]))

	srv := ftpserver.NewFtpServer(
		&FTPServer{
			Settings: &ftpserver.Settings{
				ListenAddr: "0.0.0.0:7021",
				// Use single stream for data connections
				// DisableActiveMode: true,
			},
			FileSystem: afero.Fs(fs),
		},
	)

	// Handle SIGINT and SIGTERM.
	// make chan
	sig := make(chan os.Signal, 1)
	go func() {
		<-sig
		srv.Stop()
	}()
	signal.Notify(sig, os.Interrupt)

	srv.Logger = logrus.New()
	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}

	// // server
	// ln, err := net.Listen("tcp", "0.0.0.0:7080")
	// if err != nil {
	// 	panic(err)
	// }
	// defer ln.Close()

	// fmt.Println("Starting server...")
	// err = Serve(ln, fs)
	// if err != nil {
	// 	fmt.Println("ERROR:", err)
	// }

	fs.Unmount("0:")
	fatfs.UnregisterBlockDevice(0)
	img.Close()

	fmt.Println("Done!")
}
