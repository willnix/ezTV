package ezTV

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// Eztv errors
var (
	ErrEpisodeNotFound = errors.New("episode not found")
	ErrShowNotFound    = errors.New("show not found")
	ErrEmptyResponse   = errors.New("empty response from server")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrMissingArgument = errors.New("missing argument")
)

// ShowSimple represents a show when in a list
type ShowSimple struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ShowDetails represents a show when in a list
type ShowDetails struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Cover string `json:"cover"`
	// [SxxExx][1080p|720p|hdtv]
	Episodes map[string]map[string]EpisodeSimple `json:"episodes"`
}

// EpisodeSimple represents a episode when in a list
type EpisodeSimple struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Magnet string `json:"magnet"`
}

// EpisodeDetails represents a episode when in a list
type EpisodeDetails struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"url"`
	Cover       string  `json:"cover"`
	Magnet      string  `json:"magnet"`
	Ratio       float32 `json:"ratio"`
}

// SearchShow finds shows with a title containg the keyword
// Returns error if no show is found
func SearchShow(keyword string) (*[]ShowSimple, error) {
	if keyword == "" {
		return nil, ErrMissingArgument
	}

	doc, err := goquery.NewDocument("https://eztv.ag/showlist/")
	if err != nil {
		return nil, err
	}

	showsRaw := doc.Find(fmt.Sprintf("a.thread_link:contains('%s')", keyword))
	if showsRaw.Length() < 1 {
		return nil, ErrEmptyResponse
	}

	shows := make([]ShowSimple, showsRaw.Length())

	showsRaw.Each(func(i int, s *goquery.Selection) {
		shows[i] = ShowSimple{}
		shows[i].Title = s.Text()
		shows[i].URL, _ = s.Attr("href")
	})

	if len(shows) == 0 {
		return nil, ErrShowNotFound
	}

	return &shows, nil

}

// GetShowDetails returns details of a show including a list of episodes
func GetShowDetails(URL string) (*ShowDetails, error) {
	if URL == "" {
		return nil, ErrMissingArgument
	}
	doc, err := goquery.NewDocument(fmt.Sprintf("https://eztv.ag%s", URL))
	if err != nil {
		return nil, err
	}

	// Regex for the episode ID
	epidRe := regexp.MustCompile("S[0-9]{2}E[0-9]{2}")
	// Regex for episode quality
	qualityRe := regexp.MustCompile("[0-9]{3,4}p")
	episodes := make(map[string]map[string]EpisodeSimple)

	doc.Find("tr.forum_header_border").Each(func(i int, s *goquery.Selection) {
		s.Children().Find("a").Each(func(i int, s *goquery.Selection) {
			info := s.Text()
			// It's the link to the episode details page
			// Extract url, EpID and quality from this one
			if info != "" {
				href, ok := s.Attr("href")
				if ok {
					epid := epidRe.FindString(info)
					quality := qualityRe.FindString(info)
					if quality == "" {
						quality = "hdtv"
					}
					if episodes[epid] != nil {
						episodes[epid][quality] = EpisodeSimple{Title: info, URL: href}
					} else {
						episodes[epid] = map[string]EpisodeSimple{quality: {Title: info, URL: href}}
					}
				}
				// It's the magnet link, use EpID and quality to set magnet link
			} else if s.HasClass("magnet") {
				info, ok := s.Attr("title")
				if !ok {
					return
				}
				href, ok := s.Attr("href")
				if !ok {
					return
				}
				epid := epidRe.FindString(info)
				quality := qualityRe.FindString(info)
				if quality == "" {
					quality = "hdtv"
				}
				// Episode already exists in map use existing title and url
				if (episodes[epid][quality].Title != "") && (episodes[epid][quality].URL != "") {
					episodes[epid][quality] = EpisodeSimple{
						Title:  episodes[epid][quality].Title,
						URL:    episodes[epid][quality].URL,
						Magnet: href}
				} else {
					episodes[epid] = map[string]EpisodeSimple{
						quality: {
							Title:  "",
							URL:    "",
							Magnet: href}}
				}
			}
		})
	})

	// Get show title
	s := doc.Find("b > span[itemprop='name']")
	title := s.Text()

	// Get show cover
	s = doc.Find(".show_info_main_logo > img:nth-child(1)")
	cover, _ := s.Attr("src")

	return &ShowDetails{
			Title:    title,
			URL:      URL,
			Cover:    cover,
			Episodes: episodes},
		nil
}

// GetEpisodeDetails returns details of a episode
func GetEpisodeDetails(URL string) (*EpisodeDetails, error) {
	return nil, nil
}
