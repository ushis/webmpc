package webmpc

import (
  "code.google.com/p/go.net/websocket"
  "encoding/json"
  "net/http"
)

type Server struct {
  mpd     *Mpd             // connection to MPD
  watcher *Watcher         // listens to MPD for changes
  conns   map[*Conn]bool   // websocket client pool
  drop    chan *Conn       // channel to notify the server about dropped clients
  done    chan bool        // channel to command the server to stop
  Log     chan interface{} // log channel
}

// Returns a fresh server.
//
// mpdAddr is the address of the mpd server.
// Use an empty mpdPasswd, if the mpd server doesn't require a password.
func New(mpdAddr, mpdPasswd string) (s *Server) {
  s = &Server{
    conns: make(map[*Conn]bool),
    drop:  make(chan *Conn),
    done:  make(chan bool),
    Log:   make(chan interface{}),
  }
  s.mpd = NewMpd(mpdAddr, mpdPasswd, s.Log)
  s.watcher = NewWatcher(mpdAddr, mpdPasswd, s.mpd.cmd, s.Log)
  go s.run()
  return
}

// Starts the server.
func (s *Server) run() {
  defer s.shutdown()

  for {
    select {
    case r := <-s.mpd.res:
      s.broadcast(r)
    case conn := <-s.drop:
      s.dropConn(conn)
    case <-s.done:
      return
    }
  }
}

// Stops the server.
func (s *Server) Close() {
  s.done <- true
}

// Returns the http.Handler
func (s *Server) Handler() http.Handler {
  return websocket.Handler(s.handleConnection)
}

// Handles incoming websocket connections.
func (s *Server) handleConnection(ws *websocket.Conn) {
  conn := NewConn(ws, s.mpd.cmd, s.drop, s.Log)
  s.conns[conn] = true
  conn.Run()
}

// Broadcasts a mpd result.
func (s *Server) broadcast(r *Result) {
  msg, err := json.Marshal(r)
  r.Free()

  if err != nil {
    s.Log <- err
    return
  }

  for conn, _ := range s.conns {
    conn.msg <- msg
  }
}

// Drops a connection from the pool.
func (s *Server) dropConn(conn *Conn) {
  if s.conns[conn] {
    delete(s.conns, conn)
    close(conn.msg)
  }
}

// Drops all clients, closes all channels and the connection to the mpd server.
func (s *Server) shutdown() {
  for conn, _ := range s.conns {
    s.dropConn(conn)
  }
  s.watcher.Close()
  s.mpd.Close()
  close(s.done)
  close(s.Log)
  close(s.drop)
  close(s.mpd.cmd)
}
