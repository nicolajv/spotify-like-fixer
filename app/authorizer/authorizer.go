package authorizer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/zmb3/spotify/v2"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var auth *spotifyauth.Authenticator
var ch chan *spotify.Client
var state string

func Authorize() *spotify.Client {
	godotenv.Load(".env")

	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(os.Getenv(("SPOTIFY_ID"))),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserFollowRead,
			spotifyauth.ScopeUserFollowModify,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserLibraryModify,
		),
	)
	ch = make(chan *spotify.Client)
	state = "abc123"

	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	err := exec.Command("open", url).Run()
	if err != nil {
		log.Fatal(err)
	}

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.DisplayName)
	fmt.Println("")

	return client
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok), spotify.WithRetry(true))
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}
