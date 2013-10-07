package main

import (
  "flag"
  "fmt"
  "net"
  "net/http"
  "os"
  "os/signal"
  "syscall"
  "webmpc"
)

var (
  addr      string
  index     string
  mpdAddr   string
  mpdPasswd string
)

func init() {
  flag.StringVar(&addr, "listen", ":8080", "address or socket to listen to")
  flag.StringVar(&index, "index", "./index.html", "path to index.html")
  flag.StringVar(&mpdAddr, "addr", "127.0.0.1:6600", "address of the mpd server")
  flag.StringVar(&mpdPasswd, "passwd", "", "mpd password")
}

func main() {
  flag.Parse()

  s := webmpc.New(mpdAddr, mpdPasswd)
  defer s.Close()

  l, err := listen(addr)

  if err != nil {
    die(err)
  }
  defer l.Close()

  log("Listening on:", addr)

  mux := http.NewServeMux()
  mux.HandleFunc("/", serveIndex)
  mux.Handle("/ws", s.Handler())

  go chnLog(s.Log)
  go http.Serve(l, mux)

  sig := make(chan os.Signal)
  signal.Notify(sig, syscall.SIGINT)
  <-sig
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, index)
}

func listen(addr string) (net.Listener, error) {
  if len(addr) > 0 && addr[0] == '/' {
    return net.Listen("unix", addr)
  }
  return net.Listen("tcp", addr)
}

func log(msg ...interface{}) {
  fmt.Fprintln(os.Stderr, msg...)
}

func chnLog(chn <-chan interface{}) {
  for msg := range chn {
    log(msg)
  }
}

func die(msg ...interface{}) {
  log(msg...)
  os.Exit(1)
}
