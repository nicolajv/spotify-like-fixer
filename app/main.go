package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nicolajv/spotify-like-fixer/app/authorizer"
	"github.com/zmb3/spotify/v2"
)

var client *spotify.Client

func main() {
	client = authorizer.Authorize()

	likedTracks := getLikedTracks()

	failureList := []string{}

	replaced := 0

	for _, track := range likedTracks {
		searchString := prepareQuery(track)

		result, err := client.Search(context.Background(), searchString, spotify.SearchTypeTrack)
		if err != nil || len(result.Tracks.Tracks) == 0 || track.Name != result.Tracks.Tracks[0].Name || track.Album.Name != result.Tracks.Tracks[0].Album.Name {
			failureMessage := fmt.Sprintf("Failed to search up %s by %s", track.Name, track.Artists[0].Name)
			fmt.Println(failureMessage)
			fmt.Println(searchString)
			// TODO Spotify has an issue with queries over 100 characters
			// https://community.spotify.com/t5/Spotify-for-Developers/V1-API-Search-q-query-parameter-appears-to-have-a-limit-of-100/td-p/5398898
			if len(searchString) > 100 {
				fmt.Println("Query string is too long, continuing")
				fmt.Println("")
				continue
			}
			failureList = append(failureList, failureMessage)
			fmt.Println("")
			continue
		}

		if result.Tracks.Tracks[0].ID == track.ID {
			continue
		}

		fmt.Println(track.Name, "-", track.Artists[0].Name)
		fmt.Println("No match!", "Liked id:", track.ID, "Searched id:", result.Tracks.Tracks[0].ID)

		unlikeTrack(track.ID)
		likeTrack(result.Tracks.Tracks[0].ID)
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

func prepareQuery(track spotify.SavedTrack) string {
	result := fmt.Sprintf("track:\"%s\"artist:\"%s\"album:\"%s\"", track.Name, track.Artists[0].Name, track.Album.Name)
	result = strings.Replace(result, "'", "", -1)
	result = strings.Replace(result, "(", "", -1)
	result = strings.Replace(result, ")", "", -1)
	return result
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
