package webmpc

import (
  "code.google.com/p/go.net/websocket"
  "errors"
  "github.com/fhs/gompd/mpd"
  "io"
  "net/http"
  "time"
)

type Server struct {
  mpd    *mpd.Client      // Connection to the mpd server
  addr   string           // Address of the mpd server
  passwd string           // mpd password
  conns  map[*Conn]bool   // Websocket client pool
  drop   chan *Conn       // Channel to notify the server about dropped clients
  cmd    chan *Cmd        // Command channel
  result chan *Result     // Result channel
  stop   chan bool        // Channel to command the server to stop
  err    chan error       // Error channel
  Log    chan interface{} // Log channel
}

// Returns a fresh server.
//
// mpdAddr is the address of the mpd server.
// Use an empty mpdPasswd, if the mpd server doesn't require a password.
func New(mpdAddr, mpdPasswd string) *Server {
  return &Server{
    addr:   mpdAddr,
    passwd: mpdPasswd,
    conns:  make(map[*Conn]bool),
    cmd:    make(chan *Cmd),
    result: make(chan *Result),
    drop:   make(chan *Conn),
    stop:   make(chan bool),
    err:    make(chan error, 100),
    Log:    make(chan interface{}, 100),
  }
}

// Starts the server.
func (s *Server) Run() {
  if err := s.dial(); err != nil {
    s.Log <- err
    s.shutdown()
    return
  }

  for {
    select {
    case cmd := <-s.cmd:
      go s.exec(cmd)
    case r := <-s.result:
      s.broadcast(r)
    case conn := <-s.drop:
      s.dropConn(conn)
    case err := <-s.err:
      if err = s.handleError(err); err != nil {
        s.Log <- err
        s.shutdown()
        return
      }
    case <-s.stop:
      s.shutdown()
      return
    case <-time.After(10 * time.Second):
      if err := s.mpd.Ping(); err != nil {
        s.err <- err
      }
    }
  }
}

// Stops the server.
func (s *Server) Stop() {
  s.stop <- true
}

// Returns the http.Handler
func (s *Server) Handler() http.Handler {
  return websocket.Handler(s.handleConnection)
}

// Handles incoming websocket connections.
func (s *Server) handleConnection(ws *websocket.Conn) {
  conn := NewConn(ws, s.cmd, s.drop, s.Log)
  s.conns[conn] = true
  conn.Run()
}

// Handles errors.
func (s *Server) handleError(err error) error {
  if err != io.EOF {
    s.Log <- err
    return nil
  }
  s.Log <- "lost connection to mpd"
  return s.dial()
}

// Connects to the mpd server.
func (s *Server) dial() (err error) {
  for {
    s.Log <- "connecting to mpd at: " + s.addr

    if s.mpd, err = mpd.DialAuthenticated("tcp", s.addr, s.passwd); err == nil {
      s.Log <- "connected"
      return
    }
    s.Log <- err
    s.Log <- "retrying in 3 seconds..."

    select {
    case <-s.stop:
      return errors.New("couldn't connect to mpd")
    case <-time.After(3 * time.Second):
      // Retry after 3 seconds.
    }
  }
}

// Executes a mpd command.
func (s *Server) exec(cmd *Cmd) {
  defer cmd.Free()

  if r, err := cmd.Exec(s.mpd); err != nil {
    s.err <- err
  } else {
    s.result <- r
  }
}

// Broadcasts a mpd result.
func (s *Server) broadcast(r *Result) {
  for conn, _ := range s.conns {
    conn.result <- r
  }
}

// Drops a connection from the pool.
func (s *Server) dropConn(conn *Conn) {
  if s.conns[conn] {
    delete(s.conns, conn)
    close(conn.result)
  }
}

// Drops all clients, closes all channels and the connection to the mpd server.
func (s *Server) shutdown() {
  for conn, _ := range s.conns {
    s.dropConn(conn)
  }
  close(s.result)
  close(s.stop)
  close(s.Log)
  close(s.drop)
  s.mpd.Close()
}
