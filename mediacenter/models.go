package mediacenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type VolumeState struct {
	Active             bool                     `json:"active"`
	Mute               bool                     `json:"mute"`
	Volume             float64                  `json:"volume"`
	Minimum            float64                  `json:"min"`
	Maximum            float64                  `json:"max"`
	Steps              float64                  `json:"steps"`
	ChangeCapabilities VolumeChangeCapabilities `json:"changecapabilities"`
}

func (v VolumeState) Validate() error {
	if v.Volume < v.Minimum || v.Volume > v.Maximum {
		return fmt.Errorf("Volume is %f, but has to be between %f and %f (min, max)", v.Volume, v.Minimum, v.Maximum)
	}
	return nil
}

type VolumeChangeCapabilities struct {
	Mute   bool `json:"mute"`
	UpDown bool `json:"updown"`
	Set    bool `json:"set"`
}

type SetPlaybackState string

func (s SetPlaybackState) Validate() error {
	switch s {
	case "play":
		return nil
	case "pause":
		return nil
	case "stop":
		return nil
	case "previous":
		return nil
	case "next":
		return nil
	default:
		return fmt.Errorf("Playback state '%s' is not valid", s)
	}
}

type SetSpeed int

type SeekPosition int

type SetOption struct {
	Option string
	State  bool
}

func (s SetOption) Validate() error {
	switch s.Option {
	case "random":
		return nil
	case "repeat":
		return nil
	default:
		return fmt.Errorf("Playback option '%s' is not valid", s.Option)
	}
}

type Kind string

type Play struct {
	Kind Kind
	What interface{}
}

func (p Play) Validate() error {
	switch p.Kind {
	case "url":
		if w, ok := p.What.(PlayItemURL); ok {
			return w.Validate()
		}
		return fmt.Errorf("Wrong item specified for playing %s, should be PlayItemURI but is %v", p.Kind, p.What)
	case "movie":
		if w, ok := p.What.(PlayItemMovie); ok {
			return w.Validate()
		}
		return fmt.Errorf("Wrong item specified for playing %s, should be PlayItemMovie but is %v", p.Kind, p.What)
	case "episode":
		if w, ok := p.What.(PlayItemEpisode); ok {
			return w.Validate()
		}
		return fmt.Errorf("Wrong item specified for playing %s, should be PlayItemEpisode but is %v", p.Kind, p.What)
	case "playlist":
		if w, ok := p.What.(PlayItemPlaylist); ok {
			return w.Validate()
		}
		return fmt.Errorf("Wrong item specified for playing %s, should be PlayItemPlaylist but is %v", p.Kind, p.What)
	case "artist":
		if w, ok := p.What.(PlayItemArtist); ok {
			return w.Validate()
		}
		return fmt.Errorf("Wrong item specified for playing %s, should be PlayItemArtist but is %v", p.Kind, p.What)
	case "album":
		if w, ok := p.What.(PlayItemAlbum); ok {
			return w.Validate()
		}
		return fmt.Errorf("Wrong item specified for playing %s, should be PlayItemAlbum but is %v", p.Kind, p.What)
	case "song":
		if w, ok := p.What.(PlayItemSong); ok {
			return w.Validate()
		}
		return fmt.Errorf("Wrong item specified for playing %s, should be PlayItemSong but is %v", p.Kind, p.What)
	default:
		return fmt.Errorf("Playing '%s' is not valid", p.Kind)
	}
}

type PlayItemURL string

func (p PlayItemURL) Validate() error {
	if p == "" {
		return errors.New("Playing an empty URL doesn't work")
	}
	return nil
}

type PlayItemMovie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
}

func (p PlayItemMovie) Validate() error {
	if p.Title == "" || p.Year == 0 {
		return errors.New("Playing a movie requires the title and year")
	}
	return nil
}

type PlayItemEpisode struct {
	Show    string `json:"show"`
	Season  int    `json:"season"`
	Episode int    `json:"episode"`
}

func (p PlayItemEpisode) Validate() error {
	if p.Show == "" || p.Season == 0 || p.Episode == 0 {
		return errors.New("Playing an episode requires the show, season and episode")
	}
	return nil
}

type PlayItemPlaylist string

func (p PlayItemPlaylist) Validate() error {
	if p == "" {
		return errors.New("Playing a playlist requires the name")
	}
	return nil
}

type PlayItemArtist struct {
	Artist string `json:"artist"`
}

func (p PlayItemArtist) Validate() error {
	if p.Artist == "" {
		return errors.New("Playing songs of an artist requires the artist")
	}
	return nil
}

type PlayItemAlbum struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

func (p PlayItemAlbum) Validate() error {
	if p.Artist == "" || p.Album == "" {
		return errors.New("Playing songs of an album requires the artist and album")
	}
	return nil
}

type PlayItemSong struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Song   string `json:"song"`
}

func (p PlayItemSong) Validate() error {
	if p.Artist == "" || p.Album == "" || p.Song == "" {
		return errors.New("Playing a song requires the artist, album and song")
	}
	return nil
}

type Playback struct {
	Source             string                      `json:"source"`
	State              string                      `json:"state"`
	Type               string                      `json:"type,omitempty"`
	StartTime          *PlaybackTime               `json:"starttime,omitempty"`
	EndTime            *PlaybackTime               `json:"endtime,omitempty"`
	Elapsed            PlaybackDuration            `json:"elapsed,omitempty"`
	Duration           PlaybackDuration            `json:"duration,omitempty"`
	Speed              int                         `json:"speed,omitempty"`
	AvailbleSpeeds     []int                       `json:"availablespeeds,omitempty"`
	Item               *PlaybackItem               `json:"item,omitempty"`
	PreviousItem       *PlaybackItem               `json:"previous,omitempty"`
	NextItem           *PlaybackItem               `json:"next,omitempty"`
	ChangeCapabilities *PlaybackChangeCapabilities `json:"changecapabilities,omitempty"`
	Options            *PlaybackOptions            `json:"options,omitempty"`
}

func (p Playback) Validate() error {
	if p.Source == "" || p.State == "" {
		return errors.New("Playback requires a source and state")
	}

	return nil
}

type PlaybackTime time.Time

func NewPlaybackTime(t time.Time) *PlaybackTime {
	p := PlaybackTime(t)

	return &p
}

func (p PlaybackTime) MarshalJSON() ([]byte, error) {
	s := p.Time().Format(time.RFC3339)

	return json.Marshal(s)
}

func (p *PlaybackTime) UnmarshalJSON(b []byte) error {
	s := ""
	if err := json.Unmarshal(b, &s); err != nil {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*p = PlaybackTime(t)

	return nil
}

func (p PlaybackTime) Time() time.Time {
	return time.Time(p)
}

type PlaybackDuration time.Duration

func (p PlaybackDuration) MarshalJSON() ([]byte, error) {
	d := time.Duration(p)
	millis := int(d.Nanoseconds() / 1e6)

	return json.Marshal(millis)
}

func (p *PlaybackDuration) UnmarshalJSON(b []byte) error {
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	*p = PlaybackDuration(time.Duration(i) * time.Millisecond)

	return nil
}

func (p PlaybackDuration) Duration() time.Duration {
	return time.Duration(p)
}

type PlaybackChangeCapabilities struct {
	Speed   bool `json:"speed"`
	Move    bool `json:"move"`
	Repeat  bool `json:"repeat"`
	Rotate  bool `json:"rotate"`
	Seek    bool `json:"seek"`
	Shuffle bool `json:"shuffle"`
	Zoom    bool `json:"zoom"`
}

type PlaybackOptions struct {
	Repeat string `json:"repeat"`
	Random bool   `json:"random"`
}

type PlaybackItem struct {
	Title    string               `json:"title"`
	Type     string               `json:"type,omitempty"`
	Filename string               `json:"filename,omitempty"`
	LiveTV   *PlaybackItemLiveTv  `json:"livetv,omitempty"`
	Movie    *PlaybackItemMovie   `json:"movie,omitempty"`
	Episode  *PlaybackItemEpisode `json:"episode,omitempty"`
	Song     *PlaybackItemSong    `json:"song,omitempty"`
}

type PlaybackItemLiveTv struct {
	Channel TvChannel `json:"channel"`
}

type PlaybackItemMovie struct {
	IMDBNumber    string      `json:"imdbnumber,omitempty"`
	OriginalTitle string      `json:"originaltitle,omitempty"`
	Year          int         `json:"year,omitempty"`
	Rating        *ItemRating `json:"rating,omitempty"`
}

type PlaybackItemEpisode struct {
	ShowTitle  string      `json:"showtitle"`
	Season     int         `json:"season"`
	Episode    int         `json:"episode"`
	FirstAired *time.Time  `json:"firstaired,omitempty"`
	IMDBNumber string      `json:"imdbnumber,omitempty"`
	Year       int         `json:"year,omitempty"`
	Rating     *ItemRating `json:"rating,omitempty"`
}

type PlaybackItemSong struct {
	Album  string `json:"album"`
	Artist string `json:"artist"`
	Track  int    `json:"track"`
	Total  int    `json:"total"`
	Year   int    `json:"year"`
}

type TvChannel struct {
	Type   string `json:"type"`
	Number int    `json:"number"`
	Name   string `json:"name"`
}

type ItemRating struct {
	Rating float32 `json:"rating"`
	Votes  int64   `json:"votes"`
}
