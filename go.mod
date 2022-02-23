module github.com/hrfee/go-scrobble

go 1.15

replace github.com/hrfee/go-scrobble/mpdclient => ./mpdclient

require (
	github.com/adrg/xdg v0.2.3
	github.com/hrfee/go-scrobble/mpdclient v0.0.0-00010101000000-000000000000
	github.com/pkg/browser v0.0.0-20201207095918-0426ae3fba23
	github.com/shkh/lastfm-go v0.0.0-20191215035245-89a801c244e0
	github.com/smartystreets/goconvey v1.7.2 // indirect
	gopkg.in/ini.v1 v1.62.0
)
