### go-scrobble-mpd

Same as normal go-scrobble but directly interacts with MPD only.


```bash
$ go get github.com/hrfee/go-scrobble@mpd

$ go-scrobble -help
Usage of go-scrobble:
  -config string
    	Path to config file (default "~/.config/go-scrobble-mpd.ini")
  -debug
    	Debug logging.
```

```ini
[mpd]
; Connection protocol. should be tcp or unix.
protocol = tcp
; Address of MPD.
address  = localhost:6600
; optional MPD password.
password = 

[api]
; Last.FM API Key and Secret. Generate at https://www.last.fm/api/account/create
key    =
secret = 

[general]
; How often per second to poll for track position. You can probably leave this alone.
poll-rate      = 1
; Strip features (e.g (feat. X)) from track name and artist before sending to server. This may lead to better matches.
strip-features = true
```

#### License

MIT, see LICENSE.
