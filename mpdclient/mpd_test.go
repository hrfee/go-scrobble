package mpris2client

import (
	"fmt"
	"testing"
	"time"
)

func TestRefresh(t *testing.T) {
	p, err := NewPlayer("tcp", "localhost:6600", "", 1)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}
	for {
		err = p.Refresh()
		if err != nil {
			t.Errorf("Failed to refresh: %v", err)
		}
		fmt.Printf("Title: %s\nby %s (%s)\nFrom %s\nTrack %d\n%d/%d\nPlaying: %t\n", p.Title, p.Artist, p.AlbumArtist, p.Album, p.TrackNumber, (p.Position / 1000000), p.Length, p.Playing)
		time.Sleep(time.Second)
	}
}
