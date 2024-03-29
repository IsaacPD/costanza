package youtube

import (
	"fmt"
	"os/exec"
)

type runs []struct {
	Text string
}

type videoRenderer struct {
	Title struct {
		Runs runs
	}
	OwnerText struct {
		Runs runs
	}
	LengthText struct {
		SimpleText string
	}
	VideoId string
}

type contents struct {
	Contents struct {
		TwoColumnSearchResultsRenderer struct {
			PrimaryContents struct {
				SectionListRenderer struct {
					Contents []struct {
						ItemSectionRenderer struct {
							Contents []struct {
								VideoRenderer videoRenderer
							}
						}
					}
				}
			}
		}
	}
}

type youtubeTrack struct {
	Title    string
	Uploader string
	ID       string
	Length   string

	formatID string
	cmd      *exec.Cmd
	next     *exec.Cmd
}

func (vr videoRenderer) toYoutubeTrack() *youtubeTrack {
	return &youtubeTrack{
		Title:    vr.Title.Runs[0].Text,
		Uploader: vr.OwnerText.Runs[0].Text,
		ID:       vr.VideoId,
		Length:   vr.LengthText.SimpleText,
	}
}

func (yt youtubeTrack) String() string {
	return fmt.Sprintf("Title: %s, Duration: %s, Uploader: %s, ID: %s", yt.Title, yt.Length, yt.Uploader, yt.ID)
}
