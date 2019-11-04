package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"9fans.net/go/acme"
	"github.com/fhs/gompd/v2/mpd"
)

var (
	fServer   = flag.String("server", "localhost:6600", "MPD server and port")
	fPassword = flag.String("password", "", "MPD password")

	client *mpd.Client
	mtx    sync.Mutex
)

func main() {
	var err error
	if *fPassword == "" {
		client, err = mpd.Dial("tcp", *fServer)
	} else {
		client, err = mpd.DialAuthenticated("tcp", *fServer, *fPassword)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	w, err := acme.New()
	if err != nil {
		log.Fatal(err)
	}
	w.Fprintf("tag", "Prev Pause UnPause Play Next")
	w.Name("/mpd/")

	w.Ctl("clean")

	dieChan := make(chan bool)

	// This is the function to update the display whenever something
	// changes. It also pings to keep the connection alive.
	go func() {
		wtch, err := mpd.NewWatcher("tcp", "localhost:6600", "")
		if err != nil {
			log.Fatal(err)
		}
		tckr := time.Tick(5 * time.Second)
		updateDisplay(w)
		for {
			select {
			case evt := <-wtch.Event:
				// if we display playlist events, we can get a lot of flickering
				if evt == "playlist" {
					continue
				}
				updateDisplay(w)
			case _ = <-tckr:
				// mpd will disconnect us unless we check in
				client.Ping()
			case <-dieChan:
				return
			}
		}
	}()

	for e := range w.EventChan() {
		switch e.C2 {
		case 'x', 'X': // execute
			switch string(e.Text) {
			case "Del":
				w.Ctl("delete")
				return
			case "Next":
				client.Next()
			case "Prev":
				client.Previous()
			case "Pause":
				client.Pause(true)
			case "UnPause":
				client.Pause(false)
			case "Play":
				client.Play(-1)
			case "Random":
				client.Random(!getStatusAttrBool("random"))
			case "Repeat":
				client.Repeat(!getStatusAttrBool("repeat"))
			case "Playlist":
				go playlistWindow()
			default:
				// just run the command
				w.WriteEvent(e)
			}
		case 'l', 'L':
			w.Ctl("clean")
			// Search for the string
			go searchWindow(string(e.Text))
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	close(dieChan)
}

func updateDisplay(w *acme.Win) {
	mtx.Lock()
	defer mtx.Unlock()
	w.Clear()
	status, err := client.Status()
	if err != nil {
		w.Fprintf("body", "Couldn't query mpd service: %v", err)
		return
	}

	attrs, err := client.CurrentSong()
	if err != nil {
		w.Fprintf("body", "Couldn't query mpd service: %v", err)
		return
	}

	// Put the basic controls at the top so they're always handy
	pauseString := "Pause"
	if status["state"] == "pause" {
		pauseString = "UnPause"
	}
	w.Fprintf("body", "[Prev] [%v] [Play] [Next]\n", pauseString)
	w.Fprintf("body", "[Random] [Repeat] [Playlist]\n")

	var state string
	switch status["state"] {
	case "pause":
		state = "Paused"
	case "play":
		state = "Playing"
	case "stop":
		state = "Stopped"
	}
	w.Fprintf("body", "%v: %v - %v (\"%v\")\n", state, attrs["Artist"], attrs["Title"], attrs["Album"])
	w.Fprintf("body", "Random: %v, Repeat: %v\n", status["random"], status["repeat"])
	w.Ctl("clean")
}

func getStatusAttr(key string) (string, error) {
	status, err := client.Status()
	if err != nil {
		return "", err
	}
	return status[key], nil
}

func getStatusAttrBool(key string) bool {
	s, _ := getStatusAttr(key)
	if s == "1" {
		return true
	}
	return false
}

func searchWindow(query string) {
	fields := strings.Fields(query)
	if len(fields) == 0 {
		return
	}
	w, err := acme.New()
	if err != nil {
		fmt.Printf("couldn't create new acme window: %v\n", err)
		return
	}
	w.Name("/mpd/Search")
	w.Ctl("clean")

	var songs []mpd.Attrs
	if cmd := fields[0]; cmd == "artist" || cmd == "title" || cmd == "album" {
		// they specified a command to search
		args := []string{fields[0], strings.TrimSpace(strings.TrimPrefix(query, fields[0]))}
		songs, err = client.Search(args...)
		if err != nil {
			w.Fprintf("body", "%v\n", err)
			return
		}
	} else {
		// Do a broad search
		doSearch := func(existing []mpd.Attrs, t string, f string) ([]mpd.Attrs, error) {
			s, err := client.Search(t, f)
			return append(existing, s...), err
		}
		songs, err = doSearch(songs, "artist", query)
		if err != nil {
			w.Fprintf("body", "%v\n", err)
			return
		}
		songs, err = doSearch(songs, "title", query)
		if err != nil {
			w.Fprintf("body", "%v\n", err)
			return
		}
		songs, err = doSearch(songs, "album", query)
		if err != nil {
			w.Fprintf("body", "%v\n", err)
			return
		}
	}

	var contents string
	for i, s := range songs {
		contents += fmt.Sprintf("%d/ %v - %v [%v]\n", i, s["Artist"], s["Title"], s["Album"])
	}

	w.Fprintf("body", contents)
	w.Ctl("clean")

	for e := range w.EventChan() {
		switch e.C2 {
		case 'x', 'X': // execute
			txt := string(e.Text)
			if txt == "Del" {
				w.Ctl("delete")
				return
			} else {
				w.WriteEvent(e)
			}
		case 'l', 'L': // look
			w.Ctl("clean")
			// Extract numbers
			sel := e.Text
			if len(sel) == 0 {
				sel = []byte(w.Selection())
			}
			scanner := bufio.NewScanner(bytes.NewBuffer(sel))
			for scanner.Scan() {
				// grab the ID from the line
				fields := strings.Fields(scanner.Text())
				if len(fields) == 0 {
					continue
				}
				idx, err := strconv.Atoi(strings.Trim(fields[0], "/"))
				if err != nil {
					continue
				}
				if idx < 0 || idx > len(songs)-1 {
					return
				}
				song := songs[idx]
				client.Add(song["file"])
			}
		}
	}
}

func playlistWindow() {
	w, err := acme.New()
	if err != nil {
		fmt.Printf("couldn't create new acme window: %v\n", err)
		return
	}
	w.Name("/mpd/CurrentPlaylist")
	w.Ctl("clean")
	defer w.CloseFiles()

	w.Fprintf("tag", "Clear Reload WriteBack SavePlaylist")

	populatePlaylist := func(w *acme.Win) {
		w.Clear()
		songs, err := client.PlaylistInfo(-1, -1)
		if err != nil {
			w.Fprintf("body", "Couldn't fetch playlist info: %v", err)
			return
		}
		var contents string
		for _, s := range songs {
			contents += fmt.Sprintf("%v %v - %v [%v]\n", s["Id"], s["Artist"], s["Title"], s["Album"])
		}
		w.Fprintf("body", contents)
		w.Ctl("clean")
	}
	populatePlaylist(w)

	for e := range w.EventChan() {
		switch e.C2 {
		case 'x', 'X': // execute
			txt := string(e.Text)
			if txt == "Del" {
				w.Ctl("delete")
				return
			} else if txt == "Clear" {
				w.Clear()
				client.Clear()
			} else if txt == "Reload" {
				populatePlaylist(w)
			} else if txt == "WriteBack" {
				// Grab the full playlist
				songs, err := client.PlaylistInfo(-1, -1)
				if err != nil {
					w.Errf("Couldn't fetch playlist info: %v", err)
					continue
				}
				// Now delete every song whose ID is gone
				remaining := make(map[string]bool)
				pl, err := w.ReadAll("body")
				if err != nil {
					w.Errf("Can't read playlist body: %v", err)
					continue
				}
				scanner := bufio.NewScanner(bytes.NewBuffer(pl))
				for scanner.Scan() {
					// grab the ID from the line
					fields := strings.Fields(scanner.Text())
					if len(fields) == 0 {
						continue
					}
					remaining[fields[0]] = true
				}
				for _, s := range songs {
					if !remaining[s["Id"]] {
						id, err := strconv.Atoi(s["Id"])
						if err != nil {
							continue
						}
						client.DeleteID(id)
					}
				}
				populatePlaylist(w)
			} else if strings.HasPrefix(txt, "SavePlaylist") {
				fields := strings.Fields(txt)
				if len(fields) == 2 {
					err := client.PlaylistSave(fields[1])
					if err != nil {
						w.Errf("Couldn't save playlist %v: %v\n", fields[1], err)
					}
				} else {
					w.Errf("SavePlaylist requires playlist name as argument")
				}
			} else {
				w.WriteEvent(e)
			}
		case 'l', 'L': // look
			w.Ctl("clean")
		}
	}
	return
}