# Utilities
Util =

  # Makes an element.
  mk: (name, attrs) ->
    el = document.createElement(name)
    el[attr] = val for attr, val of attrs
    el

  # Human readable duration: minutes:seconds
  humanDuration: (seconds) ->
    sec = seconds % 60
    res = "#{(seconds - sec) / 60}:"
    res += 0 if sec < 10
    res + sec

  # Debounces a function.
  debounce: (fn, delay = 100) ->
    timeout = null

    ->
      args = arguments
      window.clearTimeout(timeout)
      timeout = window.setTimeout((=> fn.apply(@, args)), delay)

  # Stops an event.
  stopEvent: (e) ->
    e.preventDefault()
    e.stopPropagation()


# Simple context menu.
ContextMenu =

  # Callback is called when an item is clicked.
  callback: null

  # The wrapper.
  el: do ->
    div = Util.mk('div')
    div.classList.add('contextmenu')
    div.style.position = 'absolute'

    div.addEventListener 'click', (e) ->
      return if e.target.nodeName isnt 'LI'
      ContextMenu.callback(e.target.dataset.option) if ContextMenu.callback?
      Util.stopEvent(e)
      ContextMenu.hide()

    div

  # Removes the menu from the DOM.
  hide: -> try document.body.removeChild(@el)

  # Displays the menu with options at position x, y.
  show: (options, x, y, callback) ->
    ul = Util.mk('ul')

    for opt, option of options
      li = Util.mk('li', textContent: option)
      li.dataset.option = opt
      ul.appendChild(li)

    @el.innerHTML = ''
    @el.appendChild(ul)
    @el.style.left = "#{x}px"
    @el.style.top = "#{y}px"
    @callback = callback
    document.body.appendChild(@el)

# Hide context menu on random window click.
window.addEventListener 'click', -> ContextMenu.hide()


# Handles the websocket.
class Socket

  # Sets up the socket.
  constructor: ->

    # Messa ge queue.
    @queue = []

    @handlers = {CurrentSong: [], Files: [], Playlist: [], Status: []}

    # Get websocket addr from window.location.
    location = window.location
    s = if location.protocol is 'https:' then 'wss://' else 'ws://'
    @addr = s + location.host + '/ws'

    # open the connection.
    @open()

  # Opens the connecion.
  open: ->
    @socket = new WebSocket(@addr, ['soap'])

    # Process queue, when connection is ready.
    @socket.addEventListener 'open', =>
      @send(@queue.shift()) while @queue.length > 0

    # Reconnect after 5 seconds.
    @socket.addEventListener 'close', =>
      window.setTimeout((=> @open()), 5000)

    # Log errors.
    @socket.addEventListener 'error', (e) ->
      console.log('WebSocket error:', e)

    # Parse incoming messages.
    @socket.addEventListener 'message', (e) =>
      try
        @receive(JSON.parse(e.data))
      catch err
        console.log('Could not parse incoming message:', err)

  # Sends data to the server. Enqueues the message, if the socket is
  # not ready yet.
  send: (data) ->
    if @socket.readyState isnt WebSocket.OPEN
      return @queue.push(data)
    @socket.send(JSON.stringify(data))

  # Dispatches data to the handlers.
  receive: (data) ->
    if not @handlers[data.Type]?
      return console.log('Unknown data type: ', data.Type)
    fn.call(@, data.Data) for fn in @handlers[data.Type]

  # Registers a new handler.
  register: (type, callback) ->
    @handlers[type].push(callback)


# Handles the database.
class Db

  # Sets up the database.
  constructor: (selector, socket) ->
    @el = document.querySelector(selector)
    @wrap = @el.querySelector('div.wrap')
    @socket = socket

    # Handle database updates.
    @socket.register 'Files', (files) => @update(files)

    # Toggle folders on click.
    @el.addEventListener 'click', (e) ->
      if e.target.nodeName is 'SPAN' and e.target.parentNode.dataset.type is 'dir'
        e.target.parentNode.classList.toggle('active')

    # Play files on double click.
    @el.addEventListener 'dblclick', (e) =>
      if e.target.nodeName is 'SPAN' and e.target.parentNode.dataset.type is 'file'
        @socket.send(Cmd: 'Add', Uri: e.target.parentNode.dataset.name)
        Util.stopEvent(e)

    # Handle dragging.
    @el.addEventListener 'dragstart', (e) =>
      @handleDragStart(e, e.target.parentNode) if e.target.nodeName is 'SPAN'

    # Display a custom context menu.
    @el.addEventListener 'contextmenu', (e) =>
      @handleContextMenu(e, e.target.parentNode) if e.target.nodeName is 'SPAN'

    # Ask server for a fresh database.
    @socket.send(Cmd: 'GetFiles')

  # Populates the database. We are building a nested list here.
  update: (files) ->
    root = Util.mk('ul')

    for f in files
      tmp = root
      dirs = f.split('/')
      file = dirs.pop()

      for dir, i in dirs
        name = dirs.slice(0, i+1).join('/')

        if tmp.lastChild?.dataset.name is name
          tmp = tmp.lastChild.lastChild
          continue

        li = Util.mk('li')
        li.dataset.name = name
        li.dataset.type = 'dir'
        li.appendChild(Util.mk('span', textContent: dir, draggable: true))
        ul = Util.mk('ul')
        li.appendChild(ul)
        tmp.appendChild(li)
        tmp = ul

      li = Util.mk('li')
      li.dataset.name = f
      li.dataset.type = 'file'
      li.appendChild(Util.mk('span', textContent: file, draggable: true))
      tmp.appendChild(li)

    try @wrap.removeChild(@wrap.lastChild)
    @wrap.appendChild(root)

  # Returns a list of uris inside a folder or file.
  getUris: (li) ->
    return [li.dataset.name] if li.dataset.type is 'file'
    (l.dataset.name for l in li.querySelectorAll('li[data-type="file"]'))

  # Populates the transfer data with the uris inside the dragged folder/file.
  handleDragStart: (e, li) ->
    e.dataTransfer.setData('application/json', JSON.stringify(type: 'uris', data: @getUris(li)))

  # Displays a context menu to add the files inside the clicked folder to playlist.
  handleContextMenu: (e, li) ->
    ContextMenu.show {AddMulti: 'Add', SetPlaylist: 'Replace'}, e.x, e.y, (action) =>
      @socket.send(Cmd: action, Uris: @getUris(li), Pos: -1)
    Util.stopEvent(e)


# Handles the playlist.
class Playlist

  # Sets up the playlist.
  constructor: (selector, socket) ->
    @el = document.querySelector(selector)
    @wrap = @el.querySelector('div.wrap')
    @curIndex = -1
    @socket = socket

    # Handle playlist updates.
    @socket.register 'Playlist', (tracks) => @update(tracks)

    # Handles status updates.
    @socket.register 'Status', (state) => @updateCurrent(window.parseInt(state.song))

    # Clear the playlist.
    @el.querySelector('span.clear').addEventListener 'click', (e) =>
      @socket.send(Cmd: 'Clear')

    # Play track on double click.
    @el.addEventListener 'dblclick', (e) =>
      return if e.target.nodeName isnt 'TD'
      @socket.send(Cmd: 'PlayId', Id: window.parseInt(e.target.parentNode.dataset.id))
      Util.stopEvent(e)

    # Display a custom context menu.
    @el.addEventListener 'contextmenu', (e) =>
      @handleContextMenu(e, e.target.parentNode) if e.target.nodeName is 'TD'

    # Handle dragging.
    @el.addEventListener 'dragstart', (e) =>
      @handleDragStart(e, e.target) if e.target.nodeName is 'TR'

    # Handle drops.
    @el.addEventListener 'drop', (e) =>
      @handleDrop(e, e.target.parentNode)

    # Stop some drag events.
    @el.addEventListener('dragover', Util.stopEvent, false)
    @el.addEventListener('dragenter', Util.stopEvent, false)
    @el.addEventListener('dragleave', Util.stopEvent, false)

    # Ask the server for a fresh playlist.
    @socket.send(Cmd: 'PlaylistInfo')

  # Populates the playlist. We are building a table here.
  update: (tracks) ->
    root = Util.mk('table')

    for track, i in tracks
      tr = Util.mk('tr', draggable: true)
      tr.dataset.id = track.Id
      tr.dataset.name = track.file
      tr.dataset.index = i

      if i is @curIndex
        tr.classList.add('active')

      td = Util.mk('td', textContent: track.Title || track.file.split('/').pop())
      td.classList.add('title')
      tr.appendChild(td)

      for info in ['Album', 'Artist']
        td = Util.mk('td', textContent: track[info] || '-')
        td.classList.add(info.toLowerCase())
        tr.appendChild(td)

      td = Util.mk('td', textContent: Util.humanDuration(track.Time))
      td.classList.add('time')
      tr.appendChild(td)
      root.appendChild(tr)

    try @wrap.removeChild(@wrap.lastChild)
    @wrap.appendChild(root)

  # Highlights the current track.
  updateCurrent: (index) ->
    return if @curIndex is index
    @curIndex = index
    row = @el.querySelectorAll('tr')[@curIndex]
    return unless row?
    act.classList.remove('active') for act in @el.querySelectorAll('tr.active')
    row.classList.add('active')

  # Populates the transfer data with the id of the dragged track.
  handleDragStart: (e, tr) ->
    data = type: 'id', data: window.parseInt(tr.dataset.id)
    e.dataTransfer.setData('application/json', JSON.stringify(data))

  # Handles drops. Distincts between moved tracks inside the playlist and added
  # tracks from the database.
  handleDrop: (e, tr) ->
    data = JSON.parse(e.dataTransfer.getData('application/json'))
    pos = window.parseInt(tr.dataset.index || -1)

    if data.type is 'uris'
      @socket.send(Cmd: 'AddMulti', Uris: data.data, Pos: pos)
    else if data.type is 'id'
      pos = @el.querySelectorAll('tr').length - 1 if pos < 0
      @socket.send(Cmd: 'MoveId', Id: data.data, Pos: pos)

  # Displays a custom context menu.
  handleContextMenu: (e, tr) ->
    ContextMenu.show {PlayId: 'Play', DeleteId: 'Remove'}, e.x, e.y, (action) =>
      @socket.send(Cmd: action, Id: window.parseInt(tr.dataset.id))
    Util.stopEvent(e)


# Handles the player.
class Player

  # Sets up the player.
  constructor: (selector, socket) ->
    @el = document.querySelector(selector)
    @vol = @el.querySelector('#volume')
    @prog = @el.querySelector('#progress')
    @progVal = @el.querySelector('#progress-val')
    @progRem = @el.querySelector('#progress-remain')
    @curTrack = @el.querySelector('#current-track')
    @curId = -1
    statusTimeout = null

    # Handle current song udpates.
    socket.register 'CurrentSong', (track) => @updateCurrent(track)

    # Handle status updates. Trigger a new status request every second.
    socket.register 'Status', (status) =>
      window.clearTimeout(statusTimeout)
      statusTimeout = window.setTimeout((-> socket.send(Cmd: 'Status')), 1000)
      socket.send(Cmd: 'CurrentSong') if @curId isnt status.songid
      @update(status)

    # Play prev track.
    @el.querySelector('#prev').addEventListener 'click', =>
      socket.send(Cmd: 'Previous')

    # Play next track.
    @el.querySelector('#next').addEventListener 'click', =>
      socket.send(Cmd: 'Next')

    # Toggle pause/play.
    @el.querySelector('#pause').addEventListener 'click', =>
      socket.send(Cmd: 'Pause', Pause: (@el.dataset.state is 'play'))

    # Toggle random playback.
    @el.querySelector('#random').addEventListener 'click', =>
      socket.send(Cmd: 'Random', Random: (@el.dataset.random is '0'))

    # Toggle repeat mode.
    @el.querySelector('#repeat').addEventListener 'click', =>
      socket.send(Cmd: 'Repeat', Repeat: (@el.dataset.repeat is '0'))

    # Set volume.
    @vol.addEventListener 'change', Util.debounce(=>
      socket.send(Cmd: 'SetVolume', Volume: window.parseInt(@vol.value)))

    # Seek current track.
    @prog.addEventListener 'change', Util.debounce(=>
      socket.send(Cmd: 'SeekId', Id: window.parseInt(@curId), Time: window.parseInt(@prog.value)))

    # Initiate the first status update.
    socket.send(Cmd: 'Status')

  # Updates the player state.
  update: (status) ->
    @el.dataset.state = status.state
    @el.dataset.random = status.random
    @el.dataset.repeat = status.repeat
    @vol.value = status.volume
    @curId = status.songid

    if status.time?
      @updateProg.apply(@, status.time.split(':').map (i) -> window.parseInt(i))

  # Updates the progress bar.
  updateProg: (cur, max) ->
    [@prog.value, @prog.max] = [cur, max]
    @progVal.textContent = Util.humanDuration(cur)
    @progRem.textContent = "-#{Util.humanDuration(if max > cur then max - cur else 0 )}"

  # Updates the current track title.
  updateCurrent: (track) ->
    track.Title ||= track.file.split('/').pop()
    title = (track[k] for k in ['Title', 'Album', 'Artist'] when track[k])
    document.title = title.join(' - ')
    @curTrack.innerHTML = ''
    @curTrack.appendChild(Util.mk('span', textContent: info)) for info in title


# GO!
sock = new Socket()
db = new Db('#db', sock)
pl = new Playlist('#playlist', sock)
player = new Player('#player', sock)
