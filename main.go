package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/OffBroadway/fatfs/pkg/fatfs"
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

	// // server
	// ln, err := net.Listen("tcp", "127.0.0.1:7564")
	// if err != nil {
	// 	panic(err)
	// }
	// defer ln.Close()

	// var errorlog, tracelog styx.Logger
	// tracelog = log.New(os.Stderr, "", 0)
	// errorlog = log.New(os.Stderr, "", 0)

	// afero9p.NewServer(afero9p.ServerOptions{
	// 	Listener: ln,
	// 	ErrorLog: errorlog,
	// 	TraceLog: tracelog,
	// }, afero.Fs(fs))

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

	// ffs, err := affuse.New(fs)
	// if err != nil {
	// 	log.Fatalf("failed to create fs: %v", err)
	// }

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)

	// ctx, cancel := context.WithCancel(context.Background())

	// go func() {
	// 	<-c
	// 	cancel()
	// }()
	// defer ffs.Unmount()
	// if err := ffs.Mount(ctx, "/home/trevor/mnt"); err != nil {
	// 	log.Fatal("failed to mount file system")
	// 	return
	// }

	// fusefs := aferofuse.NewFuseFileSystem(fs)

	// opts := &nodefs.Options{}
	// opts.Debug = true

	// mynodefs := pathfs.NewPathNodeFs(fusefs, &pathfs.PathNodeFsOptions{
	// 	ClientInodes: true,
	// })
	// server, _, err := nodefs.MountRoot("/home/trevor/mnt", mynodefs.Root(), opts)
	// if err != nil {
	// 	log.Fatalf("Mount fail: %v\n", err)
	// }
	// log.Println("Mounted!")

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// go func() {
	// 	<-c
	// 	server.Unmount()
	// }()

	// server.Serve()

	fs.Unmount("0:")
	fatfs.UnregisterBlockDevice(0)
	img.Close()

	fmt.Println("Done!")
}
