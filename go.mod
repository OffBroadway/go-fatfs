module github.com/OffBroadway/fatfs

go 1.23.4

// replace golang.org/x/net => ./alt/golang.org/x/net
// replace github.com/hanwen/go-fuse/v2 => ./alt/github.com/hanwen/go-fuse/v2
// replace gitlab.bertha.cloud/adphi/affuse => ./alt/gitlab.bertha.cloud/adphi/affuse

require (
	github.com/fclairamb/ftpserverlib v0.25.0
	github.com/fclairamb/go-log v0.5.0
	github.com/gorilla/handlers v1.5.2
	github.com/spf13/afero v1.11.0
	golang.org/x/net v0.34.0
)

require (
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)
