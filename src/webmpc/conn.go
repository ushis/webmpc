package webmpc

import (
  "code.google.com/p/go.net/websocket"
  "io"
)

type Conn struct {
  ws     *websocket.Conn    // The websocket connection.
  cmd    chan<- *Cmd        // Command channel to the server.
  die    chan<- *Conn       // Drop channel to notify the server about a drop.
  log    chan<- interface{} // Log channel.
  result chan *Result       // Result channel from the server.
}

// Returns a fresh connection.
func NewConn(ws *websocket.Conn, cmd chan<- *Cmd, die chan<- *Conn, log chan<- interface{}) *Conn {
  return &Conn{ws, cmd, die, log, make(chan *Result, 100)}
}

// Starts the send/receive loops.
//
// Receives results from the server and sends them to the client.
// Receives commands from the client and forwards them to the server.
func (c *Conn) Run() {
  defer c.ws.Close()

  go c.receive()

  for r := range c.result {
    switch err := websocket.JSON.Send(c.ws, r); {
    case err == nil:
      // Do nothing.
    case err == io.EOF:
      c.drop()
      return
    default:
      c.log <- err
    }
  }
}

// Receives commands from the client and forwards them to the server.
func (c *Conn) receive() {
  defer c.drop()

  for {
    cmd := NewCmd()

    switch err := websocket.JSON.Receive(c.ws, cmd); {
    case err == nil:
      c.cmd <- cmd
    case err == io.EOF:
      return
    default:
      c.log <- err
    }
  }
}

// Notifies the server about the dropped client.
func (c *Conn) drop() {
  c.die <- c
}
