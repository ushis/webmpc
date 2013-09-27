package webmpc

import (
  "github.com/fhs/gompd/mpd"
  "io"
  "time"
)

// Represents the connection to MPD.
type Mpd struct {
  addr   string             // address of MPD
  passwd string             // password for MPD
  err    chan error         // error channel to handle internal errors
  cmd    chan *Cmd          // command channel
  res    chan *Result       // results channel
  done   chan bool          // channel to stop the run loop
  log    chan<- interface{} // log channel
}

// Returns a new MPD and starts listening for commands.
func NewMpd(addr, passwd string, log chan<- interface{}) (m *Mpd) {
  m = &Mpd{
    addr:   addr,
    passwd: passwd,
    log:    log,
    cmd:    make(chan *Cmd),
    res:    make(chan *Result),
    err:    make(chan error),
    done:   make(chan bool),
  }
  go m.run()
  return
}

// Dials and listens for commands.
func (m *Mpd) run() {
  conn, err := m.dial()

  if err != nil {
    m.log <- err
    return
  }
  defer conn.Close()

  for {
    select {
    case cmd := <-m.cmd:
      go m.handleCmd(cmd, conn)
    case err = <-m.err:
      if err == io.EOF {
        go m.run()
        return
      }
      m.log <- err
    case <-m.done:
      return
    case <-time.After(5 * time.Second):
      go m.ping(conn)
    }
  }
}

// Opens the connection to MPD.
func (m *Mpd) dial() (c *mpd.Client, err error) {
  for {
    if c, err = mpd.DialAuthenticated("tcp", m.addr, m.passwd); err == nil {
      return
    }
    m.log <- err

    select {
    case <-m.done:
      return
    case <-time.After(time.Second):
      // continue
    }
  }
}

// Executes a command and forwards the results.
func (m *Mpd) handleCmd(cmd *Cmd, c *mpd.Client) {
  res, err := cmd.Exec(c)
  cmd.Free()

  if err != nil {
    m.err <- err
    return
  }

  if res != nil {
    m.res <- res
  }
}

// Pings MPD.
func (m *Mpd) ping(c *mpd.Client) {
  if err := c.Ping(); err != nil {
    m.err <- err
  }
}

// Stops listening for commands and closes the connection.
func (m *Mpd) Close() {
  m.done <- true
  close(m.done)
  close(m.err)
  close(m.res)
}
