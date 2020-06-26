spot-dupes

This project requires the Go wrapper for the Spotify Web API, available here: 
https://github.com/zmb3/spotify



A song is determined to be a duplicate if the track name and artist name are identical. Songs that match this criteria, but have differing Spotify URIs are still considered to be the same song.
Currently, typing "all" does not work and will exit the program instead