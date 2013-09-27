package webmpc

type Result struct {
  Type string      // The data type
  Data interface{} // The data
}

// Returns a fresh result.
func NewResult(t string, data interface{}) *Result {
  return &Result{t, data}
}

type StoredPlaylist struct {
  Name   string      // Name of the playlist
  Tracks interface{} // the playlists tracks
}

// Returns a fresh stored playlist.
func NewStoredPlaylist(name string, tracks interface{}) *StoredPlaylist {
  return &StoredPlaylist{name, tracks}
}
