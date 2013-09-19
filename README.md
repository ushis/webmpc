# Webmpc

Realtime [mpd](http://www.musicpd.org/) client running in the browser.

![Screenshot](http://i.imgur.com/7qMURyV.png)

Here is a video: https://www.youtube.com/watch?v=ZH-GvE2ZYrM

## How

Webmpc connects to a server via Websockets, that forwards the commands to mpd
and broadcasts the results to all connected clients.

## Get it

To compile the server you need a [Go](http://golang.org/) compiler and a
[Ruby](https://www.ruby-lang.org) interpreter to compile the client. The
Makefile uses [godag](https://code.google.com/p/godag/) as compiler front-end.

Get the dependencies.

```
go get github.com/fhs/gompd/mpd
go get code.google.com/p/net.go/websocket
gem install coffee-script uglifier sass
```

Build the binaries.

```
git clone https://github.com/ushis/webmpc
cd webmpc
make
```

Start the server.

```
./webmpc -listen=:8080
```

Visit [localhost:8080](http://localhost:8080) in your browser of choice.


## License

```
The MIT License (MIT)

Copyright (c) 2013 ushi <ushi@honkgong.info>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
