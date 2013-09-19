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
  if err := conn.Add(cmd.Uri); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
}

//
func addId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if _, err := conn.AddId(cmd.Uri, cmd.Pos); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
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

  if err := list.End(); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
}

//
func clear(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.Clear(); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
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
  if err := conn.Delete(cmd.Start, cmd.End); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
}

//
func delId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.DeleteId(cmd.Id); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
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
  if err := conn.MoveId(cmd.Id, cmd.Pos); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
}

//
func next(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.Next(); err != nil {
    return nil, err
  }
  return currentSong(cmd, conn)
}

//
func pause(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.Pause(cmd.Pause); err != nil {
    return nil, err
  }
  return status(cmd, conn)
}

//
func play(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.Play(cmd.Pos); err != nil {
    return nil, err
  }
  return currentSong(cmd, conn)
}

//
func playId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.PlayId(cmd.Id); err != nil {
    return nil, err
  }
  return currentSong(cmd, conn)
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
  if err := conn.Previous(); err != nil {
    return nil, err
  }
  return currentSong(cmd, conn)
}

//
func random(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.Random(cmd.Random); err != nil {
    return nil, err
  }
  return status(cmd, conn)
}

//
func repeat(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.Repeat(cmd.Repeat); err != nil {
    return nil, err
  }
  return status(cmd, conn)
}

//
func seek(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.Seek(cmd.Pos, cmd.Time); err != nil {
    return nil, err
  }
  return currentSong(cmd, conn)
}

//
func seekId(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.SeekId(cmd.Id, cmd.Time); err != nil {
    return nil, err
  }
  return currentSong(cmd, conn)
}

//
func setPlaylist(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  list := conn.BeginCommandList()
  list.Clear()

  for _, uri := range cmd.Uris {
    list.Add(uri)
  }
  list.Play(0)

  if err := list.End(); err != nil {
    return nil, err
  }
  return playlistInfo(cmd, conn)
}

//
func setVolume(cmd *Cmd, conn *mpd.Client) (*Result, error) {
  if err := conn.SetVolume(cmd.Volume); err != nil {
    return nil, err
  }
  return status(cmd, conn)
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
  if err := conn.Stop(); err != nil {
    return nil, err
  }
  return status(cmd, conn)
}
