package youtube

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/anaskhan96/soup"
	"github.com/isaacpd/costanza/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	searchURLFormat = "http://youtube.com/results?search_query=%s"
	initialDataRE   = regexp.MustCompile("(window\\[\"ytInitialData\"]|var ytInitialData)\\s*=\\s*(.*);")
)

func Search(query string) []youtubeTrack {
	query = strings.Join(strings.Split(query, " "), "+")
	url := fmt.Sprintf(searchURLFormat, query)

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(url)
	err := util.DoWithRedirects(req, res)
	if err != nil {
		logrus.Warnf("Error getting youtube search results: %s", err)
		return []youtubeTrack{}
	}
	return extractSearchResults(res.Body())
}

func extractSearchResults(body []byte) []youtubeTrack {
	doc := soup.HTMLParse(string(body))

	var elements []soup.Root
	root := doc.Find("ytd-search")
	err := root.Error
	if root.Error == nil {
		root = root.Find("ytd-item-section-renderer")
		err = root.Error
		if root.Error == nil {
			elements = root.FindAll("a", "id", "video-title")
		}
	}
	if err != nil {
		logrus.Warnf("Error getting html tags: %s", err)
	}
	logrus.Tracef("Video Title elements: %v", elements)
	var results []youtubeTrack
	for _, e := range elements {
		results = append(results,
			youtubeTrack{
				ID:    e.Attrs()["href"],
				Title: e.Attrs()["title"],
			})
	}
	if len(results) == 0 {
		return extractPolymerResults(body)
	}
	return results
}

func extractPolymerResults(body []byte) []youtubeTrack {
	res := initialDataRE.Find(body)
	if res == nil {
		logrus.Warnf("Initial data regular expression didn't match anything")
		return []youtubeTrack{}
	}

	j := util.FindJson(string(res))

	var jsonOut contents
	err := json.Unmarshal([]byte(j), &jsonOut)
	if err != nil {
		logrus.Warnf("Couldn't parse json of initial data %s", err)
		return []youtubeTrack{}
	}

	var results []youtubeTrack
	for _, c := range jsonOut.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.SectionListRenderer.Contents[0].ItemSectionRenderer.Contents {
		vr := c.VideoRenderer
		if vr.VideoId == "" {
			continue
		}
		results = append(results, vr.toYoutubeTrack())
	}
	return results
}
