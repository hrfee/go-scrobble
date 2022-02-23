package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	mpd "github.com/hrfee/go-scrobble/mpdclient"
	"github.com/pkg/browser"
	"github.com/shkh/lastfm-go/lastfm"
	"gopkg.in/ini.v1"
)

var (
	configFile = filepath.Join(xdg.ConfigHome, "go-scrobble-mpd.ini")
	poll       = 1
	debug      = false
	stripFeat  = false
)

type trackDetails struct {
	title, artist, albumArtist, album string
	trackNumber                       int
	length                            int
	playing                           bool
	started                           time.Time
}

func stripFeatures(s string) string {
	lower := strings.ToLower(s)
	for _, v := range []string{"(feat", "( feat", "feat."} {
		i := strings.Index(lower, v)
		if i != -1 {
			if s[i-1] == ' ' {
				return s[:i-1]
			}
			return s[:i]
		}
	}
	return s
}

func (t *trackDetails) update(p *mpd.Player) {
	t.title = p.Title
	t.artist = p.Artist
	t.albumArtist = p.AlbumArtist
	t.album = p.Album
	t.trackNumber = p.TrackNumber
	t.playing = p.Playing
	t.length = p.Length
}

func (t *trackDetails) equals(o trackDetails) bool {
	return t.title == o.title && t.artist == o.artist && t.albumArtist == o.albumArtist && t.album == o.album
}

func (t *trackDetails) isDuplicateScrobble(o trackDetails) bool {
	if !t.equals(o) {
		return false
	}
	// If the new scrobble started after the calculated end time of the last scrobble, its not a duplicate
	if t.started.After(o.started.Add(time.Duration(o.length) * time.Second)) {
		return false
	}
	return true
}

func genParams(p *mpd.Player) (map[string]interface{}, bool) {
	if p.Title == "" || p.Artist == "" || p.Album == "" {
		return nil, false
	}
	params := map[string]interface{}{
		"artist": p.Artist,
		"track":  p.Title,
		"album":  p.Album,
	}
	if stripFeat {
		params["artist"] = stripFeatures(p.Artist)
		params["track"] = stripFeatures(p.Title)
	}
	if p.AlbumArtist != "" {
		params["albumArtist"] = p.AlbumArtist
	}
	if p.Length != 0 {
		params["duration"] = p.Length
	}
	if p.TrackNumber != 0 {
		params["trackNumber"] = p.TrackNumber
	}
	return params, true
}

func validScrobble(p *mpd.Player) bool {
	pos := int(p.Position / 1000000)
	return ((float64(pos) / float64(p.Length)) > 0.5) || pos > 4*60
}

func withinTimeRange(calculated int, actual int) bool {
	return actual > (calculated-5) || actual < (calculated+5)
}

func playerInfo(p *mpd.Player) string {
	return fmt.Sprintf("Track: %s, Artist: %s, Album: %s", p.Title, p.Artist, p.Album)
}

func serverResponse(res interface{}, err error) string {
	return fmt.Sprintf("Response: %+v\nError: %+v\n", res, err)
}

func setKey(config *ini.File, section, key, value, comment string) {
	config.Section(section).Key(key).SetValue(value)
	config.Section(section).Key(key).Comment = comment
}

func genDefaultConfig() {
	dir, _ := filepath.Split(configFile)
	os.MkdirAll(dir, os.FileMode(0700))
	f, err := os.Create(configFile)
	if err != nil {
		log.Fatalf("Failed to create new config at \"%s\"", configFile)
	}
	f.Close()
	tempConfig, err := ini.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to create new config at \"%s\"", configFile)
	}
	setKey(tempConfig, "mpd", "protocol", "tcp", "Connection protocol. should be tcp or unix.")
	setKey(tempConfig, "mpd", "address", "localhost:6600", "Address of MPD.")
	setKey(tempConfig, "mpd", "password", "", "optional MPD password.")
	setKey(tempConfig, "api", "key", "", "Last.FM API Key and Secret. Generate at https://www.last.fm/api/account/create")
	setKey(tempConfig, "api", "secret", "", "")
	setKey(tempConfig, "general", "poll-rate", "1", "How often per second to poll for track position. You can probably leave this alone.")
	setKey(tempConfig, "general", "strip-features", "true", "Strip features (e.g (feat. X)) from track name and artist before sending to server. This may lead to better matches.")
	err = tempConfig.SaveTo(configFile)
	if err != nil {
		log.Fatalf("Failed to save template config at \"%s\"", configFile)
	}
	fmt.Printf("Saved template config at \"%s\". You'll need to fill in the API key and Secret. Generate these now? [yY/nN]\n>: ", configFile)
	choice := ""
	fmt.Scanln(&choice)
	if strings.ToLower(choice) == "y" {
		url := "https://www.last.fm/api/account/create"
		browser.OpenURL(url)
		fmt.Printf("%s\nFill in the form (Details are unimportant) then add the key/secret to the config file.\n", url)
	}
}

func getSessionKey(api *lastfm.Api, config *ini.File) (sessionKey string) {
	token, err := api.GetToken()
	if err != nil {
		log.Fatalln("Couldn't get token from Last.FM:", err)
	}
	url := api.GetAuthTokenUrl(token)
	browser.OpenURL(url)
	fmt.Printf("%s\nAuthorize and then press Enter to continue.\n>: ", url)
	fmt.Scanln()
	err = api.LoginWithToken(token)
	if err != nil {
		log.Fatalln("Couldn't login:", err)
	}
	sessionKey = api.GetSessionKey()
	config.Section("api").Key("sk").SetValue(sessionKey)
	err = config.SaveTo(configFile)
	if err != nil {
		log.Fatalln("Failed to write config file:", err)
	}
	fmt.Printf("Your session key was added to \"%s\". Reauthorization is necessary if lost.\n", configFile)
	return
}

func main() {
	flag.BoolVar(&debug, "debug", debug, "Debug logging.")
	flag.StringVar(&configFile, "config", configFile, "Path to config file")
	flag.Parse()

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		genDefaultConfig()
		return
	}
	config, err := ini.Load(configFile)
	if err != nil {
		log.Fatalln("Couldn't read config:", err)
	}
	poll = config.Section("general").Key("poll-rate").MustInt(poll)
	stripFeat = config.Section("general").Key("strip-features").MustBool(stripFeat)

	key := config.Section("api").Key("key").MustString("")
	secret := config.Section("api").Key("secret").MustString("")
	if key == "" || secret == "" {
		log.Fatalln("Couldn't get API key/secret from config")
	}
	api := lastfm.New(key, secret)
	sessionKey := config.Section("api").Key("sk").MustString("")
	if sessionKey == "" {
		sessionKey = getSessionKey(api, config)
	}
	api.SetSession(sessionKey)

	protocol := config.Section("mpd").Key("protocol").MustString("tcp")
	address := config.Section("mpd").Key("address").MustString("localhost:6600")
	password := config.Section("mpd").Key("password").String()
	player, err := mpd.NewPlayer(protocol, address, password, poll)
	if err != nil {
		log.Fatalf("Failed to connect to MPD: %v", err)
	}

	last := trackDetails{}
	now := trackDetails{}
	lastScrobbled := trackDetails{}
	go player.Listen()
	for v := range player.Messages {
		if v.Name != "refresh" {
			continue
		}
		now.update(player)
		if now.equals(last) {
			continue
		}
		last = now
		if !player.Playing {
			continue
		}
		params, ok := genParams(player)
		if !ok {
			if debug {
				log.Println("Ignoring due to missing metadata")
			}
			continue
		}
		res, err := api.Track.UpdateNowPlaying(params)
		if err == nil {
			log.Println("Now Playing: " + playerInfo(player))
		}
		if debug {
			log.Println(serverResponse(res, err))
		}
		go func(p *mpd.Player, details trackDetails, params map[string]interface{}) {
			currentDetails := trackDetails{}
			pos := int(p.Position / 1000000)
			ts := time.Now().Add(time.Duration(-pos) * time.Second)
			params["timestamp"] = ts.Unix()
			for {
				time.Sleep(time.Duration(poll) * time.Second)
				pos += poll
				if !p.Exists() {
					fmt.Println("exist")
					return
				}
				p.GetPosition()
				p.Refresh()
				if !withinTimeRange(pos, int(p.Position/1000000)) {
					fmt.Println("timerange", pos, int(p.Position/1000000))
					return
				}
				currentDetails.update(p)
				if !details.equals(currentDetails) {
					return
				}
				if !validScrobble(p) {
					continue
				}
				details.started = ts
				if details.isDuplicateScrobble(lastScrobbled) {
					return
				}
				res, err := api.Track.Scrobble(params)
				msg := serverResponse(res, err)
				if err != nil {
					log.Println("Failed to scrobble: ", msg)
				}
				if res.Ignored == "1" {
					log.Println("Server ignored scrobble: ", msg)
				}
				if debug {
					log.Println("Scrobbled: ", msg)
				}
				lastScrobbled = details
				return
			}
		}(player, now, params)
	}
}
