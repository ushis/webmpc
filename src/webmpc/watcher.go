package webmpc

import (
  "github.com/fhs/gompd/mpd"
  "io"
  "time"
)

// Represents a listener, that listens to MPD's internal changes.
type Watcher struct {
  addr   string             // address of MPD
  passwd string             // MPD password
  done   chan bool          // channel to stop the run loop
  log    chan<- interface{} // log channel
  cmd    chan<- *Cmd        // channel to send commands to MPD
}

// Returns a new watcher and starts listening to changes inside MPD.
func NewWatcher(addr, passwd string, cmd chan<- *Cmd, log chan<- interface{}) (w *Watcher) {
  w = &Watcher{
    addr:   addr,
    passwd: passwd,
    cmd:    cmd,
    log:    log,
    done:   make(chan bool),
  }
  go w.run()
  return
}

// Opens a connection to MPD and starts listening for changes.
func (w *Watcher) run() {
  watcher, err := w.dial()

  if err != nil {
    w.log <- err
    return
  }
  defer watcher.Close()

  go func() {
    for event := range watcher.Event {
      w.handleEvent(event)
    }
  }()

  go func() {
    for err := range watcher.Error {
      w.handleError(err)
    }
  }()

  <-w.done
}

// Opens a connection to MPD.
func (w *Watcher) dial() (watcher *mpd.Watcher, err error) {
  for {
    if watcher, err = mpd.NewWatcher("tcp", w.addr, w.passwd); err == nil {
      return
    }

    select {
    case <-w.done:
      return
    case <-time.After(time.Second):
      // retry
    }
  }
}

// Handles a received event and may send a follow up command to MPD, to get
// the updates.
func (w *Watcher) handleEvent(subsystem string) {
  cmd := NewCmd()

  switch subsystem {
  case "mixer":
    fallthrough
  case "options":
    fallthrough
  case "player":
    cmd.Cmd = "Status"
  case "playlist":
    cmd.Cmd = "PlaylistInfo"
  case "stored_playlist":
    cmd.Cmd = "ListPlaylists"
  case "database":
    cmd.Cmd = "GetFiles"
  default:
    return
  }
  w.cmd <- cmd
}

// Handles errors.
func (w *Watcher) handleError(err error) {
  if err != io.EOF {
    w.log <- err
    return
  }
  w.done <- true
  go w.run()
}

// Stops listening to MPD and closes the connection.
func (w *Watcher) Close() {
  w.done <- true
  close(w.done)
}
