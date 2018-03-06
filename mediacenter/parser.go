package mediacenter

import (
	"encoding/json"
	"strconv"
)

func ParseVolumeState(b []byte) (VolumeState, error) {
	v := VolumeState{}

	if err := json.Unmarshal(b, &v); err != nil {
		return v, err
	}
	if err := v.Validate(); err != nil {
		return v, err
	}

	return v, nil
}

func ParseSetPlaybackState(b []byte) (SetPlaybackState, error) {
	s := SetPlaybackState(string(b))
	if err := s.Validate(); err != nil {
		return SetPlaybackState(""), err
	}

	return s, nil
}

func ParseSetOption(option string, b []byte) (SetOption, error) {
	s := SetOption{
		Option: option,
	}

	state, err := strconv.ParseBool(string(b))
	if err != nil {
		return s, err
	}
	s.State = state
	if err := s.Validate(); err != nil {
		return s, err
	}

	return s, nil
}

func ParseSetSpeed(b []byte) (SetSpeed, error) {
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}

	return SetSpeed(i), nil
}

func ParseSeekPosition(b []byte) (SeekPosition, error) {
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}

	return SeekPosition(i), nil
}

func ParsePlay(k Kind, b []byte) (Play, error) {
	p := Play{
		Kind: k,
	}

	var err error
	switch p.Kind {
	case "url":
		p.What = PlayItemURL(string(b))
	case "movie":
		w := PlayItemMovie{}
		err = json.Unmarshal(b, &w)
		p.What = w
	case "episode":
		w := PlayItemEpisode{}
		err = json.Unmarshal(b, &w)
		p.What = w
	case "playlist":
		p.What = PlayItemPlaylist(string(b))
	case "artist":
		w := PlayItemArtist{}
		err = json.Unmarshal(b, &w)
		p.What = w
	case "album":
		w := PlayItemAlbum{}
		err = json.Unmarshal(b, &w)
		p.What = w
	case "song":
		w := PlayItemSong{}
		err = json.Unmarshal(b, &w)
		p.What = w
	}

	if err != nil {
		return p, err
	}
	if err := p.Validate(); err != nil {
		return p, err
	}

	return p, nil
}

func ParsePlayback(b []byte) (Playback, error) {
	p := Playback{}

	if err := json.Unmarshal(b, &p); err != nil {
		return p, err
	}
	if err := p.Validate(); err != nil {
		return p, err
	}

	return p, nil
}
