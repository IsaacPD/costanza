package spotify

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/zmb3/spotify"
)

var (
	auth   spotify.Authenticator
	client spotify.Client
)

const (
	state       = "discord"
	redirectURL = "http://localhost:8080/spotify/callback"
)

func Init() {
	// the redirect URL must be an exact match of a URL you've registered for your application
	// scopes determine which permissions the user is prompted to authorize
	os.Setenv("SPOTIFY_ID", strings.TrimSpace(os.Getenv("SPOTIFY_ID")))
	os.Setenv("SPOTIFY_SECRET", strings.TrimSpace(os.Getenv("SPOTIFY_SECRET")))
	scopes := []string{
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserModifyPlaybackState,
	}
	auth = spotify.NewAuthenticator(redirectURL, scopes...)
}

func AuthenticateSpotify(c cmd.Context) {
	// get the user to this URL - how you do that is up to you
	// you should specify a unique state string to identify the session
	url := auth.AuthURL("discord")
	c.Send(fmt.Sprintf("Allow costanza access to your spotify at %s", url))
}

func SpotifyRedirectHandler(w http.ResponseWriter, r *http.Request) {
	// use the same state string here that you used to generate the URL
	token, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}
	// create a client using the specified token
	client = auth.NewClient(token)
	fmt.Fprint(w, "Success: Received the token 😁")
}

func GetCurrentlyPlayingSong() (*spotify.CurrentlyPlaying, error) {
	return client.PlayerCurrentlyPlaying()
}

func Next() error {
	return client.Next()
}
