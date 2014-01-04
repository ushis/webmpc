package webmpc

type Result struct {
  Type string      // The data type
  Data interface{} // The data
}

var resultPool = make(chan *Result, 100)

// Returns a fresh result.
func NewResult(t string, data interface{}) (r *Result) {
  select {
  case r = <-resultPool:
    // Got one from the pool.
  default:
    r = new(Result)
  }
  r.Type = t
  r.Data = data
  return
}

//
func (r *Result) Free() {
  select {
  case resultPool <- r:
    // Stored it in the pool.
  default:
    // Pool is full. It's a job for the GC.
  }
}

type StoredPlaylist struct {
  Name   string      // Name of the playlist
  Tracks interface{} // the playlists tracks
}

// Returns a fresh stored playlist.
func NewStoredPlaylist(name string, tracks interface{}) *StoredPlaylist {
  return &StoredPlaylist{name, tracks}
}
