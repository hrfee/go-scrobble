package mpris2client

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/fhs/gompd/mpd"
)

// Various paths and values to use elsewhere.
const (
	INTERFACE = "org.mpris.MediaPlayer2"
	PATH      = "/org/mpris/MediaPlayer2"
	// For the NameOwnerChanged signal.
	MATCH_NOC = "type='signal',path='/org/freedesktop/DBus',interface='org.freedesktop.DBus',member='NameOwnerChanged'"
	// For the PropertiesChanged signal. It doesn't match exactly (couldn't get that to work) so we check it manually.
	MATCH_PC    = "type='signal',path='/org/mpris/MediaPlayer2',interface='org.freedesktop.DBus.Properties'"
	EventAdd    = "add"
	EventRemove = "remove"
)

var knownPlayers = map[string]string{
	"plasma-browser-integration": "Browser",
	"noson":                      "Noson",
}

var knownBrowsers = map[string]string{
	"mozilla":  "Firefox",
	"chrome":   "Chrome",
	"chromium": "Chromium",
}

type Message struct {
	Name, Value string
}

// Player represents an active media player.
type Player struct {
	network, addr, password           string
	Title, Artist, AlbumArtist, Album string
	TrackNumber, Length               int // -1 when track number unavailable
	Position                          int64
	Playing, Stopped                  bool
	conn                              *mpd.Client
	poll                              int
	Messages                          chan Message
}

// NewPlayer returns a new player object.
func NewPlayer(network, addr, password string, poll int) (p *Player, err error) {
	p = &Player{
		network:  network,
		addr:     addr,
		password: password,
		Messages: make(chan Message),
		poll:     poll,
	}
	if password != "" {
		p.conn, err = mpd.DialAuthenticated(network, addr, password)
		if err != nil {
			return
		}
	} else {
		p.conn, err = mpd.Dial(network, addr)
		if err != nil {
			return
		}
	}
	return
}

func (p *Player) Close() error {
	return p.conn.Close()
}

func (p *Player) Reconnect() error {
	err := p.conn.Close()
	if err != nil {
		return err
	}
	if p.password != "" {
		p.conn, err = mpd.DialAuthenticated(p.network, p.addr, p.password)
	} else {
		p.conn, err = mpd.Dial(p.network, p.addr)
	}
	return err
}

func (p *Player) Exists() bool {
	return p.conn.Ping() == nil
}

func (p *Player) String() string {
	return fmt.Sprintf("Title: %s, Playing: %t", p.Title, p.Playing)
}

// Refresh grabs playback info.
func (p *Player) Refresh() (err error) {
	status, err := p.conn.Status()
	if err != nil {
		return err
	}
	// fmt.Print("STATUS:\n\n")
	// for f, v := range status {
	// 	fmt.Printf("%v: %v\n", f, v)
	// }

	switch status["state"] {
	case "pause":
		p.Stopped = false
		p.Playing = false
	case "play":
		p.Stopped = false
		p.Playing = true
	default: // "stop
		p.Stopped = true
		p.Playing = false
	}

	length, err := strconv.ParseFloat(status["duration"], 64)
	if err != nil {
		length = -1
	}
	p.Length = int(length)
	pos, err := strconv.ParseFloat(status["elapsed"], 64)
	if err != nil {
		pos = 0
	}
	p.Position = int64(pos * 1000000)

	current, err := p.conn.CurrentSong()
	if err != nil {
		return err
	}
	// fmt.Print("CURRENT:\n\n")
	// for f, v := range current {
	// 	fmt.Printf("%v: %v\n", f, v)
	// }

	p.Title = current["Title"]
	p.Artist = current["Artist"]
	p.AlbumArtist = current["AlbumArtist"]
	p.Album = current["Album"]
	trackNum, err := strconv.ParseInt(current["Track"], 10, 64)
	if err != nil {
		trackNum = -1
	}
	p.TrackNumber = int(trackNum)

	return nil
	// val, err := p.Player.GetProperty(INTERFACE + ".Player.PlaybackStatus")
	// if err != nil {
	// 	p.Playing = false
	// 	p.Stopped = false
	// 	p.Metadata = map[string]dbus.Variant{}
	// 	p.Title = ""
	// 	p.Artist = ""
	// 	p.AlbumArtist = ""
	// 	p.Album = ""
	// 	p.TrackNumber = -1
	// 	p.Length = 0
	// 	return
	// }
	// strVal := val.String()
	// if strings.Contains(strVal, "Playing") {
	// 	p.Playing = true
	// 	p.Stopped = false
	// } else if strings.Contains(strVal, "Paused") {
	// 	p.Playing = false
	// 	p.Stopped = false
	// } else {
	// 	p.Playing = false
	// 	p.Stopped = true
	// }
	// metadata, err := p.Player.GetProperty(INTERFACE + ".Player.Metadata")
	// if err != nil {
	// 	p.Metadata = map[string]dbus.Variant{}
	// 	p.Title = ""
	// 	p.Artist = ""
	// 	p.AlbumArtist = ""
	// 	p.Album = ""
	// 	p.TrackNumber = -1
	// 	p.Length = 0
	// 	return
	// }
	// p.Metadata = metadata.Value().(map[string]dbus.Variant)
	// switch artist := p.Metadata["xesam:artist"].Value().(type) {
	// case []string:
	// 	p.Artist = strings.Join(artist, ", ")
	// case string:
	// 	p.Artist = artist
	// default:
	// 	p.Artist = ""
	// }
	// switch albumArtist := p.Metadata["xesam:albumArtist"].Value().(type) {
	// case []string:
	// 	p.AlbumArtist = strings.Join(albumArtist, ", ")
	// case string:
	// 	p.AlbumArtist = albumArtist
	// default:
	// 	p.AlbumArtist = ""
	// }
	// switch title := p.Metadata["xesam:title"].Value().(type) {
	// case string:
	// 	p.Title = title
	// default:
	// 	p.Title = ""
	// }
	// switch album := p.Metadata["xesam:album"].Value().(type) {
	// case string:
	// 	p.Album = album
	// default:
	// 	p.Album = ""
	// }
	// switch trackNumber := p.Metadata["xesam:trackNumber"].Value().(type) {
	// case int32:
	// 	p.TrackNumber = int(trackNumber)
	// default:
	// 	p.TrackNumber = -1
	// }
	// switch length := p.Metadata["mpris:length"].Value().(type) {
	// case int64:
	// 	p.Length = int(length) / 1000000
	// case uint64:
	// 	p.Length = int(length) / 1000000
	// default:
	// 	p.Length = 0
	// }
	// return nil
}

func µsToString(µs int64) string {
	seconds := int(µs / 1000000)
	minutes := int(seconds / 60)
	seconds -= minutes * 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (p *Player) GetPosition() bool {
	data, err := p.conn.Status()
	if err != nil {
		return false
	}
	pos, err := strconv.ParseFloat(data["elapsed"], 64)
	if err != nil {
		p.Position = 0
		return false
	}
	p.Position = int64(pos * 1000000)
	return true
}

// Returns value instead of writing it.
func (p *Player) getPosition() (int64, bool) {
	data, err := p.conn.Status()
	if err != nil {
		return 0, false
	}
	pos, err := strconv.ParseFloat(data["elapsed"], 64)
	if err != nil {
		return 0, false
	}
	return int64(pos * 1000000), true
}

func (p *Player) Listen() {
	for {
		prevTitle := p.Title
		prevArtist := p.Artist
		err := p.Refresh()
		if err != nil {
			connected := false
			for i := 0; i < 10; i++ {
				err = p.Reconnect()
				connected = err == nil
				if connected {
					break
				}
				time.Sleep(15 * time.Second)
			}
			if !connected {
				log.Fatalf("Failed to reconnect to MPD: %v", err)
			}
			p.Refresh()
		}
		if p.Title != prevTitle || p.Artist != prevArtist {
			p.Messages <- Message{Name: "refresh", Value: ""}
		}
		time.Sleep(time.Duration(p.poll) * time.Second)
	}
}
