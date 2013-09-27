package webmpc

import (
  "errors"
  "github.com/fhs/gompd/mpd"
)

// Map of mpd commands.
var commands = map[string]func(*Cmd, *mpd.Client) (*Result, error){
  "Add":          add,
  "AddId":        addId,
  "AddMulti":     addMulti,
  "Clear":        clear,
  "CurrentSong":  currentSong,
  "Delete":       del,
  "DeleteId":     delId,
  "GetFiles":     getFiles,
  "MoveId":       moveId,
  "Next":         next,
  "Pause":        pause,
  "Play":         play,
  "PlayId":       playId,
  "PlaylistInfo": playlistInfo,
  "Previous":     previous,
  "Random":       random,
  "Repeat":       repeat,
  "Seek":         seek,
  "SeekId":       seekId,
  "SetPlaylist":  setPlaylist,
  "SetVolume":    setVolume,
  "Status":       status,
  "Stop":         stop,
}

// Pool of free Cmd structs.
var cmdPool = make(chan *Cmd, 100)

// A Cmd is combination of a command and different arguments.
type Cmd struct {
  Cmd    string
  Uri    string
  Uris   []string
  Id     int
  Pause  bool
  Pos    int
  Random bool
  Repeat bool
  Start  int
  End    int
  Time   int
  Volume int
}

// Returns a new command.
func NewCmd() (c *Cmd) {
  select {
  case c = <-cmdPool:
    // Got one from the pool.
  default:
    c = new(Cmd)
  }
  return
}

// Frees a command.
func (c *Cmd) Free() {
  select {
  case cmdPool <- c:
    // Stored it in the pool.
  default:
    // Pool is full. It's a job for the gc.
  }
}

// Executes the command.
func (c *Cmd) Exec(conn *mpd.Client) (*Result, error) {
  if fn, ok := commands[c.Cmd]; ok {
    return fn(c, conn)
  }
  return nil, errors.New("Unknown command: " + c.Cmd)
}

//
func add(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Add(cmd.Uri)
  return nil, err
}

//
func addId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  _, err := conn.AddId(cmd.Uri, cmd.Pos)
  return nil, err
}

//
func addMulti(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  list := conn.BeginCommandList()

  for i, uri := range cmd.Uris {
    if cmd.Pos < 0 {
      list.Add(uri)
    } else {
      list.AddId(uri, cmd.Pos+i)
    }
  }
  return nil, list.End()
}

//
func clear(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Clear()
  return nil, err
}

//
func currentSong(_ *Cmd, conn *mpd.Client) (r *Result, err error) {
  info, err := conn.CurrentSong()

  if err == nil {
    r = NewResult("CurrentSong", info)
  }
  return
}

//
func del(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Delete(cmd.Start, cmd.End)
  return nil, err
}

//
func delId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.DeleteId(cmd.Id)
  return nil, err
}

//
func getFiles(_ *Cmd, conn *mpd.Client) (r *Result, err error) {
  files, err := conn.GetFiles()

  if err == nil {
    r = NewResult("Files", files)
  }
  return
}

//
func moveId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.MoveId(cmd.Id, cmd.Pos)
  return nil, err
}

//
func next(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Next()
  return nil, err
}

//
func pause(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Pause(cmd.Pause)
  return nil, err
}

//
func play(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Play(cmd.Pos)
  return nil, err
}

//
func playId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.PlayId(cmd.Id)
  return nil, err
}

//
func playlistInfo(_ *Cmd, conn *mpd.Client) (r *Result, err error) {
  pl, err := conn.PlaylistInfo(-1, -1)

  if err == nil {
    r = NewResult("Playlist", pl)
  }
  return
}

//
func previous(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Previous()
  return nil, err
}

//
func random(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Random(cmd.Random)
  return nil, err
}

//
func repeat(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Repeat(cmd.Repeat)
  return nil, err
}

//
func seek(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Seek(cmd.Pos, cmd.Time)
  return nil, err
}

//
func seekId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.SeekId(cmd.Id, cmd.Time)
  return nil, err
}

//
func setPlaylist(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  list := conn.BeginCommandList()
  list.Clear()

  for _, uri := range cmd.Uris {
    list.Add(uri)
  }
  list.Play(0)

  return nil, list.End()
}

//
func setVolume(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.SetVolume(cmd.Volume)
  return nil, err
}

//
func status(cmd *Cmd, conn *mpd.Client) (r *Result, err error) {
  state, err := conn.Status()

  if err == nil {
    r = NewResult("Status", state)
  }
  return
}

//
func stop(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  err := conn.Stop()
  return nil, err
}
