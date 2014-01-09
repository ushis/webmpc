(function() {
  'use strict';

  // Utilities
  var Util = {

    // Makes an element.
    mk: function(name, attrs) {
      var el = document.createElement(name);
      Util.extend(el, attrs);
      return el;
    },

    // Extends an object with the properties of another.
    extend: function(dst, src) {
      for (var k in src) {
        if (src.hasOwnProperty(k)) {
          dst[k] = src[k];
        }
      }
    },

    // Debounces a function.
    debounce: function(fn, delay) {
      var timeout = null;

      if (delay === null || delay === undefined) {
        delay = 100;
      }

      return function() {
        var that = this, args = arguments;
        window.clearTimeout(timeout);

        timeout = window.setTimeout(function() {
          fn.apply(that, args);
        }, delay);
      };
    },

    // Stops an event.
    stopEvent: function(e) {
      e.stopPropagation();
      e.preventDefault();
    },

    // Human readable duration: minutes:seconds
    humanDuration: function(seconds) {
      if (!seconds) {
        return '0:00';
      }
      var sec = seconds % 60;
      var res = ((seconds - sec) / 60) + ':';

      if (sec < 10) {
        res += '0';
      }
      return res + sec;
    }
  };


  // Thin wrapper around localStorage.
  var Store = {

    // Well... a cache.
    cache: {},

    // Stores a key value pair.
    set: function(key, val) {
      this.cache[key] = val;

      try {
        window.localStorage.setItem(key, JSON.stringify(val));
      } catch (e) {
        console.debug('Store.set: localStorage.setItem:', e);
      }
    },

    // Returns a value or the given fallback or null.
    get: function(key, fallback) {
      if (fallback === undefined) {
        fallback = null;
      }

      if (this.cache[key] !== null && this.cache[key] !== undefined) {
        return this.cache[key];
      }

      try {
        this.cache[key] = JSON.parse(window.localStorage.getItem(key));
      } catch (e) {
        console.debug('Store.get: JSON.parse:', e);
        return fallback;
      }
      return (this.cache[key] === null) ? fallback : this.cache[key];
    }
  };


  // Handles the Websocket.
  var Socket = function() {

    // The actual socket.
    this.sock = null;

    // Message queue
    this.queue = [];

    // Event handlers.
    this.handlers = {
      CurrentSong: [],
      Files: [],
      Playlist: [],
      Status: []
    };

    // Get websocket adr from window.location.
    var loc = window.location;

    if (loc.protocol === 'https://') {
      this.addr = 'wss://' + loc.host + '/ws';
    } else {
      this.addr = 'ws://' + loc.host + '/ws';
    }

    // Open the connection.
    this.open();
  };

  // Opens the connection.
  Socket.prototype.open = function() {
    var that = this;

    this.sock = new WebSocket(this.addr, ['soap']);

    // Process queue, when connection is ready.
    this.sock.addEventListener('open', function() {
      while (that.queue.length > 0) {
        that.send(that.queue.shift());
      }
    });

    // Reconnect after 5 seconds.
    this.sock.addEventListener('close', function() {
      window.setTimeout(function() {
        that.open();
      }, 5000);
    });

    // Log errors.
    this.sock.addEventListener('error', function(e) {
      console.debug('WebSocket: error:', e);
    });

    // Parse incoming messages.
    this.sock.addEventListener('message', function(e) {
      var data;

      try {
        data = JSON.parse(e.data);
      } catch (err) {
        console.debug('WebSocket: message: JSON.parse:', err);
      }
      that.receive(data);
    });
  };

  // Sends data to the server.
  Socket.prototype.send = function(data) {
    if (this.sock.readyState !== WebSocket.OPEN) {
      this.queue.push(data);
    } else {
      this.sock.send(JSON.stringify(data));
    }
  };

  // Dispatches recevied data.
  Socket.prototype.receive = function(data) {
    var handlers = this.handlers[data.Type];

    if (handlers === undefined) {
      console.debug('Socket.receive: unknown data type:', data.Type);
      return;
    }

    for (var i = 0, len = handlers.length; i < len; i++) {
      handlers[i].call(this, data.Data);
    }
  };

  // Registers a new handler.
  Socket.prototype.register = function(type, fn) {
    this.handlers[type].push(fn);
  };


  // Track database.
  var Db = function(selector, sock) {
    var that = this;

    this.el = document.querySelector(selector);
    this.wrap = this.el.querySelector('div.wrap');
    this.sock = sock;

    // Handle database updates.
    this.sock.register('Files', function(files) {
      that.update(files);
    });

    // Handles clicks.
    this.el.addEventListener('click', function(e) {
      that.handleClick(e.target);
    });

    // Handle double clicks.
    this.el.addEventListener('dblclick', function(e) {
      that.handleDblClick(e.target);
    });

    // Handle dragging.
    this.el.addEventListener('dragstart', function(e) {
      that.handleDragStart(e);
    });

    // Get a fresh
    this.sock.send({Cmd: 'GetFiles'});
  };

  //
  Db.prototype.update = function(files) {
    var root = Util.mk('ul');
    var active = Store.get('db.active', {});

    for (var i = 0, len = files.length; i < len; i++) {
      var tmp = root;
      var dirs = files[i].split('/');
      var file = dirs.pop();

      for (var j = 0, _len = dirs.length; j < _len; j++) {
        var name = dirs.slice(0, j + 1).join('/');

        if (tmp.lastChild && tmp.lastChild.dataset.name === name) {
          tmp = tmp.lastChild.lastChild;
          continue;
        }
        var li = Util.mk('li');
        li.dataset.name = name;
        li.dataset.type = 'dir';

        if (active[name]) {
          li.classList.add('active');
        }
        var span = Util.mk('span', {textContent: dirs[j], draggable: true});
        li.appendChild(span);

        var ul = Util.mk('ul');
        li.appendChild(ul);

        tmp.appendChild(li);
        tmp = ul;
      }
      var li = Util.mk('li');
      li.dataset.name = files[i];
      li.dataset.type = 'file';

      var span = Util.mk('span', {textContent: file, draggable: true});
      li.appendChild(span);

      tmp.appendChild(li);
    }

    try {
      this.wrap.removeChild(this.wrap.lastChild);
    } catch (_) {}

    this.wrap.appendChild(root);
  };

  // Returns all uris in a folder.
  Db.prototype.getUris = function(li) {
    if (li.dataset.type === 'file') {
      return [li.dataset.name];
    }
    var uris = [];
    var lis = li.querySelectorAll('li[data-type="file"]');

    for (var i = 0, len = lis.length; i < len; i++) {
      uris.push(lis[i].dataset.name);
    }
    return uris;
  };

  //
  Db.prototype.handleClick = function(el) {
    if (el.nodeName !== 'SPAN' || el.parentNode.dataset.type !== 'dir') {
      return;
    }
    el.parentNode.classList.toggle('active');
    var lis = this.el.querySelectorAll('li.active');
    var active = {};

    for (var i = 0, len = lis.length; i < len; i++) {
      active[lis[i].dataset.name] = true;
    }
    Store.set('db.active', active);
  };

  //
  Db.prototype.handleDblClick = function(el) {
    if (el.nodeName === 'SPAN' && el.parentNode.dataset.type === 'file') {
      this.sock.send({Cmd: 'Add', Uri: el.parentNode.dataset.name});
    }
  };

  //
  Db.prototype.handleDragStart = function(e) {
    var el = e.target;

    if (el.nodeName === 'SPAN') {
      var data = {type: 'uris', data: this.getUris(el.parentNode)};
      e.dataTransfer.setData('application/json', JSON.stringify(data));
    }
  };


  // Playlist
  var Playlist = function(selector, sock) {
    var that = this;
    var clicked = null;
    var dragging = false;

    this.el = document.querySelector(selector);
    this.wrap = this.el.querySelector('div.wrap');
    this.sock = sock;
    this.curId = -1;

    //
    this.sock.register('Playlist', function(tracks) {
      that.update(tracks);
    });

    //
    this.sock.register('Status', function(state) {
      that.updateCurrent(window.parseInt(state.songid));
    });

    //
    this.el.querySelector('span.clear').addEventListener('click', function() {
      that.sock.send({Cmd: 'Clear'});
    });

    //
    this.el.addEventListener('click', function(e) {
      clicked = e.target;

      // Wait a momment, in case it is a double click.
      window.setTimeout(function() {
        if (clicked === e.target) {
          that.handleClick(e.target);
        }
      }, 200);
    });

    //
    this.el.addEventListener('dblclick', function(e) {
      clicked = null;
      that.handleDblClick(e.target);
    });

    //
    this.el.addEventListener('dragstart', function(e) {
      dragging = true;
      that.handleDragStart(e);
    });

    //
    this.el.addEventListener('drop', function(e) {
      dragging = false;
      that.handleDrop(e);
    });

    //
    this.el.addEventListener('dragover', Util.stopEvent, false);
    this.el.addEventListener('dragenter', Util.stopEvent, false);
    this.el.addEventListener('dragleave', Util.stopEvent, false);

    //
    document.addEventListener('drop', function(e) {
      dragging = false;
      that.handleDropOut(e);
    });

    //
    var dragOutHandler = function(e) {
      if (dragging) {
        Util.stopEvent(e);
      }
    };

    //
    document.addEventListener('dragover', dragOutHandler, false);
    document.addEventListener('dragenter', dragOutHandler, false);
    document.addEventListener('dragleave', dragOutHandler, false);

    //
    this.sock.send({Cmd: 'PlaylistInfo'});
  };

  //
  Playlist.prototype.update = function(tracks) {
    var root = Util.mk('table');

    for (var i = 0, len = tracks.length; i < len; i++) {
      var track = tracks[i];

      var tr = Util.mk('tr', {draggable: true});
      tr.dataset.id = track.Id;
      tr.dataset.name = track.file;
      tr.dataset.index = i;

      if (window.parseInt(track.Id) === this.curId) {
        tr.classList.add('active');
      }
      var title = Util.mk('td');
      title.textContent = track.Title || track.file.split('/').pop();
      title.classList.add('title');
      tr.appendChild(title);

      var album = Util.mk('td');
      album.textContent = track.Album || '-';
      album.classList.add('album');
      tr.appendChild(album);

      var artist = Util.mk('td');
      artist.textContent = track.Artist || '-';
      artist.classList.add('artist');
      tr.appendChild(artist);

      var time = Util.mk('td');
      time.textContent = Util.humanDuration(track.Time);
      time.classList.add('time');
      tr.appendChild(time);

      root.appendChild(tr);
    }

    try {
      this.wrap.removeChild(this.wrap.lastChild);
    } catch (_) {}

    this.wrap.appendChild(root);
  };

  //
  Playlist.prototype.updateCurrent = function(id) {
    if (this.curId === id) {
      return;
    }
    var active = this.el.querySelectorAll('tr.active');

    for (var i = 0, len = active.length; i < len; i++) {
      active[i].classList.remove('active');
    }
    var row = this.el.querySelector('tr[data-id="' + id + '"]');

    if (row === null) {
      console.debug('Playlist.updateCurrent: invalid track id:', id);
      return;
    }
    row.classList.add('active');
    this.curId = id;
  };

  //
  Playlist.prototype.unselect = function(trs) {
    for (var i = 0, len = trs.length; i < len; i++) {
      trs[i].classList.remove('selected');
    }
  };

  //
  Playlist.prototype.unselectAll = function() {
    this.unselect(this.el.querySelectorAll('tr.selected'));
  };

  //
  Playlist.prototype.select = function(start, end) {
    var startIndex = window.parseInt(start.dataset.index);
    var endIndex = window.parseInt(end.dataset.index);

    if (startIndex > endIndex) {
      var tmp = start;
      start = end;
      end = tmp;
    }

    for (var el = start; el !== end; el = el.nextSibling) {
      el.classList.add('selected');
    }
    end.classList.add('selected');
  };

  //
  Playlist.prototype.length = function() {
    return this.el.querySelectorAll('tr').length;
  };

  //
  Playlist.prototype.add = function(uris, pos) {
    this.sock.send({Cmd: 'AddMulti', Uris: uris, Pos: pos});
  };

  //
  Playlist.prototype.move = function(start, end, pos) {
    // Calculate the new position, if the tracks are moved down the playlist.
    if (start < pos) {
      pos -= end - start;
    }
    this.sock.send({Cmd: 'Move', Start: start, End: end, Pos: pos});
  };

  //
  Playlist.prototype.moveId = function(id, pos) {
    var tr = this.el.querySelector('tr[data-id="' + id + '"]');

    if (tr === null) {
      console.debug('Playlist.moveId: invalid track id:', id);
      return;
    }
    var index = window.parseInt(tr.dataset.index);

    // Calculate the new position, if the track is moved down the playlist.
    if (index < pos) {
      pos -= 1;
    }
    this.sock.send({Cmd: 'MoveId', Id: id, Pos: pos});
  };

  //
  Playlist.prototype.del = function(start, end) {
    this.sock.send({Cmd: 'Delete', Start: start, End: end});
  };

  //
  Playlist.prototype.delId = function(id) {
    this.sock.send({Cmd: 'DeleteId', Id: id});
  };

  //
  Playlist.prototype.handleClick = function(el) {
    if (el.nodeName !== 'TD') {
      return;
    }
    el = el.parentNode;

    // Remove selection if the clicked element is already selected.
    if (el.classList.contains('selected')) {
      this.unselectAll();
      return;
    }
    var selected = this.el.querySelectorAll('tr.selected');

    // Select everything between the already selected and the clicked element.
    if (selected.length === 1) {
      this.select(selected[0], el);
      return;
    }

    // Select the clicked element.
    this.unselect(selected);
    el.classList.add('selected');
  };

  //
  Playlist.prototype.handleDblClick = function(el) {
    if (el.nodeName === 'TD') {
      var id = window.parseInt(el.parentNode.dataset.id);
      this.sock.send({Cmd: 'PlayId', Id: id});
    }
    this.unselectAll();
  };

  //
  Playlist.prototype.handleDragStart = function(e) {
    var el = e.target;

    if (el.nodeName !== 'TR') {
      return;
    }
    var data;

    if (el.classList.contains('selected')) {
      var els = this.el.querySelectorAll('tr.selected');
      var start = window.parseInt(els[0].dataset.index);
      var end = window.parseInt(els[els.length-1].dataset.index) + 1;
      data = {type: 'indexes', data: {start: start, end: end}};
    } else {
      data = {type: 'id', data: window.parseInt(el.dataset.id)};
    }
    e.dataTransfer.setData('application/json', JSON.stringify(data));
  };

  //
  Playlist.prototype.handleDrop = function(e) {
    var el = e.target;
    var data = JSON.parse(e.dataTransfer.getData('application/json'));
    var pos = window.parseInt(el.parentNode.dataset.index || this.length());

    switch (data.type) {
    case 'uris':
      this.add(data.data, pos);
      break;
    case 'indexes':
      this.move(data.data.start, data.data.end, pos);
      break;
    case 'id':
      this.moveId(data.data, pos);
      break;
    default:
      console.debug('Playlist.handleDrop: unexpected data type:', data.type);
    }
    Util.stopEvent(e);
  };

  //
  Playlist.prototype.handleDropOut = function(e) {
    var data = JSON.parse(e.dataTransfer.getData('application/json'));

    switch(data.type) {
    case 'indexes':
      this.del(data.data.start, data.data.end);
      break;
    case 'id':
      this.delId(data.data);
      break;
    default:
      console.debug('Playlist.handleDropOut: unexpected data type:', data.type);
    }
    Util.stopEvent(e);
  };


  //
  var Player = function(selector, sock) {
    var that = this;
    var statusTimeout = null;

    this.el = document.querySelector(selector);
    this.vol = this.el.querySelector('#volume');
    this.prog = this.el.querySelector('#progress');
    this.progVal = this.el.querySelector('#progress-val');
    this.progRem = this.el.querySelector('#progress-remain');
    this.curTrack = this.el.querySelector('#current-track');
    this.icon = Util.mk('link', {rel: 'icon', type: 'image/png'});
    this.curId = -1;
    this.sock = sock;

    //
    this.sock.register('CurrentSong', function(track) {
      that.updateCurrent(track);
    });

    //
    this.sock.register('Status', function(state) {
      window.clearTimeout(statusTimeout);

      statusTimeout = window.setTimeout(function() {
        that.sock.send({Cmd: 'Status'});
      }, 1000);

      if (that.curId !== window.parseInt(state.songid)) {
        that.sock.send({Cmd: 'CurrentSong'});
      }
      that.update(state);
    });

    //
    this.el.querySelector('#prev').addEventListener('click', function() {
      that.sock.send({Cmd: 'Previous'});
    });

    //
    this.el.querySelector('#next').addEventListener('click', function() {
      that.sock.send({Cmd: 'Next'});
    });

    //
    this.el.querySelector('#pause').addEventListener('click', function() {
      if (that.el.dataset.state === 'stop') {
        that.sock.send({Cmd: 'Play', Pos: -1});
        return;
      }
      that.sock.send({Cmd: 'Pause', Pause: that.el.dataset.state === 'play'});
    });

    //
    this.el.querySelector('#random').addEventListener('click', function() {
      that.sock.send({Cmd: 'Random', Random: that.el.dataset.random === '0'});
    });

    //
    this.el.querySelector('#repeat').addEventListener('click', function() {
      that.sock.send({Cmd: 'Repeat', Repeat: that.el.dataset.repeat === '0'});
    });

    //
    this.vol.addEventListener('change', Util.debounce(function() {
      that.sock.send({
        Cmd: 'SetVolume',
        Volume: window.parseInt(that.vol.value)
      });
    }));

    //
    this.prog.addEventListener('change', Util.debounce(function() {
      that.sock.send({
        Cmd: 'SeekId',
        Id: that.curId,
        Time: window.parseInt(that.prog.value)
      });
    }));

    //
    that.sock.send({Cmd: 'Status'});
  };

  //
  Player.prototype.update = function(state) {
    this.el.dataset.state = state.state;
    this.el.dataset.random = state.random;
    this.el.dataset.repeat = state.repeat;
    this.vol.value = state.volume;
    this.curId = window.parseInt(state.songid);

    if (state.state === 'play') {
      this.icon.href = './play.png';
    } else {
      this.icon.href = './pause.png';
    }

    try {
      document.head.removeChild(this.icon);
    } catch (_) {}

    document.head.appendChild(this.icon);

    if (state.time === null || state.time === undefined) {
      return;
    }
    var time = state.time.split(':');
    this.updateProg(window.parseInt(time[0]), window.parseInt(time[1]));
  };

  //
  Player.prototype.updateProg = function(cur, max) {
    this.prog.value = cur;
    this.prog.max = max;
    this.progVal.textContent = Util.humanDuration(cur);

    var rem = (max > cur) ? (max - cur) : 0;
    this.progRem.textContent = '-' + Util.humanDuration(rem);
  };

  //
  Player.prototype.updateCurrent = function(t) {
    if (!t.Title && t.file) {
      t.Title = t.file.split('/').pop();
    }

    var title = [t['Title'], t['Album'], t['Artist']].filter(function(s) {
      return !!s;
    });

    document.title = title.join(' - ');
    this.curTrack.innerHTML = '';

    for (var i = 0, len = title.length; i < len; i++) {
      this.curTrack.appendChild(Util.mk('span', {textContent: title[i]}));
    }
  };


  // GO!
  var sock = new Socket();
  var db = new Db('#db', sock);
  var pl = new Playlist('#playlist', sock);
  var player = new Player('#player', sock);
}).call(this);
