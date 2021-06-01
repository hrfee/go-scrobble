### go-scrobble

An ugly little MPRIS2 scrobbler for last.FM. mpris-scrobbler seemed to rarely work, and my music library naming means tracks with features in the title or artist aren't properly matched. This switches to the currently playing player automatically and strips features (optional). Uses [lastfm-go](https://github.com/shkh/lastfm-go).

The XESAM Ontology was very useful when making this, but the archive i found was super slow. I've put up a simple one [here](https://xesam.hrfee.dev) if anyone wants it.

```bash
$ go get github.com/hrfee/go-scrobble

$ go-scrobble -help
Usage of go-scrobble:
  -config string
    	Path to config file (default "~/.config/go-scrobble.ini")
  -debug
    	Debug logging.
```

```ini
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
