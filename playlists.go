//  1. Register an application at: https://developer.spotify.com/my-applications/
//       - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/zmb3/spotify"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func bigPlaylist(ID spotify.ID, client *spotify.Client, size int) *spotify.PlaylistTrackPage {
	currentList, _ := client.GetPlaylistTracks(ID)
	if size > 100 {
		for i := 100; i < size; i = i + 100 {
			opt := spotify.Options{Offset: &i}
			tracks, _ := client.GetPlaylistTracksOpt(ID, &opt, "")
			currentList.Tracks = append(currentList.Tracks, tracks.Tracks...)
		}
	}
	return currentList
}

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState,
		spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadCollaborative)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	playlists, err := client.CurrentUsersPlaylists()
	fmt.Println("total playlists: ", playlists.Total)
	allPlaylists := playlists.Playlists
	if playlists.Total > 20 {
		for i := 20; i < playlists.Total; i = i + 20 {
			opt := spotify.Options{Offset: &i}
			playlists, err = client.CurrentUsersPlaylistsOpt(&opt)
			allPlaylists = append(allPlaylists, playlists.Playlists...)
		}
		for i, s := range allPlaylists {
			fmt.Println(i, s.Name)
		}
	} else {
		for i := 0; i < len(playlists.Playlists); i++ {
			fmt.Println(i, playlists.Playlists[i].Name)
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select two playlists <1 2> or type <all>")
	playlist1, _ := reader.ReadString('\n')
	playlist1 = strings.Trim(playlist1, "\r\n")
	var matches map[string]string
	matches = make(map[string]string)
	if playlist1 == "all" {
		fmt.Println("all selected")
	} else {
		playlistsToCompare := strings.Split(playlist1, " ")
		fmt.Printf("%q\n", playlistsToCompare)
		compare := []int{}
		for i := 0; i < len(playlistsToCompare); i++ {
			n, _ := strconv.ParseFloat(playlistsToCompare[i], 10)
			compare = append(compare, int(n))
		}
		for _, n := range compare {
			//currentList, _ := client.GetPlaylistTracks(allPlaylists[n].ID)
			currentList := bigPlaylist(allPlaylists[n].ID, client, int(allPlaylists[n].Tracks.Total))
			for _, p := range currentList.Tracks {
				matches[p.Track.Name+" - "+p.Track.Artists[0].Name] += " | <" + allPlaylists[n].Name + ">"
			}
		}
		for k, v := range matches {
			if strings.Count(v, " | ") > 1 {
				fmt.Println(k, ":", v[3:])
				continue
			}
			delete(matches, v)
		}
	}
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}
