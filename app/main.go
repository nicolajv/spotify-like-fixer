package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/nicolajv/spotify-like-fixer/app/authorizer"
	"github.com/zmb3/spotify/v2"
)

var client *spotify.Client
var interactive *bool

func main() {
	interactive = flag.Bool("interactive", false, "use to perform non-exact matching, but require a prompt before replacement")
	flag.Parse()

	client = authorizer.Authorize()

	likedTracks := getLikedTracks()

	failureList := []string{}

	replaced := 0

	for _, track := range likedTracks {
		result, err := search(track)
		if err != nil {
			failureMessage := fmt.Sprintf("Failed to search up %s by %s with external id %s", track.Name, track.Artists[0].Name, track.ExternalIDs["isrc"])
			fmt.Println(failureMessage)
			fmt.Println("")
			failureList = append(failureList, failureMessage)
			continue
		}

		if result.ID == track.ID {
			continue
		}

		fmt.Println(track.Name, "-", track.Artists[0].Name)
		fmt.Println("No match!", "Liked id:", track.ID, "Searched id:", result.ID)
		fmt.Println("No match!", "Liked isrc:", track.ExternalIDs["isrc"], "Searched isrc:", result.ExternalIDs["isrc"])

		if *interactive {
			fmt.Println(track.Name, "by", track.Artists[0].Name, "from", track.Album.Name)
			fmt.Println("will be replaced with")
			fmt.Println(result.Name, "by", result.Artists[0].Name, "from", result.Album.Name)

			confirmation := WaitForConfirmation()
			if !confirmation {
				fmt.Println("")
				continue
			}
		}

		unlikeTrack(track.ID)
		likeTrack(result.ID)
		replaced++

		fmt.Println("")
	}

	fmt.Println("Failed replacements:")
	for _, failure := range failureList {
		fmt.Println(failure)
	}
	fmt.Println("")

	fmt.Println("Replacements complete! Successful replacements:", replaced)
	fmt.Println("Failed replacements:", len((failureList)))
}

func getLikedTracks() []spotify.SavedTrack {
	finalList := []spotify.SavedTrack{}

	offset := 0
	hasNext := true

	for hasNext {
		likedTracks, err := client.CurrentUsersTracks(context.Background(), spotify.Limit(50), spotify.Offset(offset))
		if err != nil {
			log.Fatal(err)
		}

		finalList = append(finalList, likedTracks.Tracks...)

		if likedTracks.Next != "" {
			offset = offset + 50
		} else {
			hasNext = false
		}
	}

	return finalList
}

func search(originalTrack spotify.SavedTrack) (*spotify.FullTrack, error) {
	results, err := client.Search(context.Background(), fmt.Sprintf("isrc:\"%s\"", originalTrack.ExternalIDs["isrc"]), spotify.SearchTypeTrack, spotify.Limit(50))
	if err != nil || len(results.Tracks.Tracks) == 0 {
		return nil, errors.New("failed to search up track")
	}

	// If something in the list has the correct ID, return that
	for _, track := range results.Tracks.Tracks {
		if originalTrack.ID == track.ID && originalTrack.ExternalIDs["isrc"] == track.ExternalIDs["isrc"] {
			return &track, nil
		}
	}

	// If we can find something that matches all parameters, return that
	for _, track := range results.Tracks.Tracks {
		if originalTrack.Name == track.Name && originalTrack.Album.Name == track.Album.Name && originalTrack.Artists[0].Name == track.Artists[0].Name && originalTrack.ExternalIDs["isrc"] == track.ExternalIDs["isrc"] {
			return &track, nil
		}
	}

	// Otherwise if something matches album + artist, return that
	for _, track := range results.Tracks.Tracks {
		if originalTrack.Album.Name == track.Album.Name && originalTrack.Artists[0].Name == track.Artists[0].Name && originalTrack.ExternalIDs["isrc"] == track.ExternalIDs["isrc"] {
			return &track, nil
		}
	}

	// Otherwise return the first result that matches just the artist
	for _, track := range results.Tracks.Tracks {
		if originalTrack.Artists[0].Name == track.Artists[0].Name && originalTrack.ExternalIDs["isrc"] == track.ExternalIDs["isrc"] {
			return &track, nil
		}
	}
	// If there is still no match, return the first item from the correct artist, as long as we are in interactive mode
	if *interactive {
		for _, track := range results.Tracks.Tracks {
			if originalTrack.Artists[0].Name == track.Artists[0].Name {
				return &track, nil
			}
		}
	}
	return nil, errors.New("failed to search up track")
}

func unlikeTrack(trackId spotify.ID) {
	fmt.Println("Unliking track:", trackId.String())
	err := client.RemoveTracksFromLibrary(context.Background(), trackId)
	if err != nil {
		log.Fatal(err)
	}
}

func likeTrack(trackId spotify.ID) {
	fmt.Println("Liking track:", trackId.String())
	err := client.AddTracksToLibrary(context.Background(), trackId)
	if err != nil {
		log.Fatal(err)
	}
}

func WaitForConfirmation() bool {
	var input string

	fmt.Printf("Is this okay? [y/n]: ")
	_, err := fmt.Scan(&input)
	if err != nil {
		panic(err)
	}

	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	if input == "y" || input == "yes" {
		return true
	}
	return false
}
