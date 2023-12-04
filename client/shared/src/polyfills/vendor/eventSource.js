// From https://github.com/EventSource/eventsource.
// Sets global `EventSource` for Node, which is required for streaming search.
// Used for VS Code web as well to be able to add Authorization header.

const original = require('original')

const parse = require('url').parse
const events = require('events')
const http = require('http')
const https = require('https')
const util = require('util')

let fixedHeaders = {}
let proxyAgent = null

module.exports = function polyfillEventSource(headers, agent) {
  fixedHeaders = { ...headers }
  proxyAgent = agent

  global.EventSource = EventSource

  // It's safe to use a browser implementation of Event and MessageEvent if we only polyfill to add
  // support for additional header fields.
  if (typeof global.MessageEvent === 'undefined') {
    global.MessageEvent = MessageEventPolyfill
  }
  if (typeof global.Event === 'undefined') {
    global.Event = EventPolyfill
  }
}

const httpsOptions = new Set([
  'pfx',
  'key',
  'passphrase',
  'cert',
  'ca',
  'ciphers',
  'rejectUnauthorized',
  'secureProtocol',
  'servername',
  'checkServerIdentity',
])

const bom = [239, 187, 191]
const colon = 58
const space = 32
const lineFeed = 10
const carriageReturn = 13

function hasBom(buf) {
  return bom.every((charCode, index) => buf[index] === charCode)
}

/**
 * Creates a new EventSource object
 *
 * @param {string} url the URL to which to connect
 * @param {Object} [eventSourceInitDict] extra init params. See README for details.
 * @api public
 **/
function EventSource(url, eventSourceInitDict) {
  let readyState = EventSource.CONNECTING
  Object.defineProperty(this, 'readyState', {
    get() {
      return readyState
    },
  })

  Object.defineProperty(this, 'url', {
    get() {
      return url
    },
  })

  const self = this
  self.reconnectInterval = 1000
  self.connectionInProgress = false

  function onConnectionClosed(message) {
    if (readyState === EventSource.CLOSED) {
      return
    }
    readyState = EventSource.CONNECTING
    _emit('error', new Event('error', { message }))

    // The url may have been changed by a temporary
    // redirect. If that's the case, revert it now.
    if (reconnectUrl) {
      url = reconnectUrl
      reconnectUrl = null
    }
    setTimeout(() => {
      if (readyState !== EventSource.CONNECTING || self.connectionInProgress) {
        return
      }
      self.connectionInProgress = true
      connect()
    }, self.reconnectInterval)
  }

  let request
  let lastEventId = ''
  if (eventSourceInitDict && eventSourceInitDict.headers && eventSourceInitDict.headers['Last-Event-ID']) {
    lastEventId = eventSourceInitDict.headers['Last-Event-ID']
    delete eventSourceInitDict.headers['Last-Event-ID']
  }

  let discardTrailingNewline = false
  let data = ''
  let eventName = ''

  var reconnectUrl = null

  function connect() {
    const options = parse(url)
    let isSecure = options.protocol === 'https:'
    options.headers = {
      Accept: 'text/event-stream',
      ...fixedHeaders,
    }
    if (lastEventId) {
      options.headers['Last-Event-ID'] = lastEventId
    }
    if (eventSourceInitDict && eventSourceInitDict.headers) {
      for (const index in eventSourceInitDict.headers) {
        const header = eventSourceInitDict.headers[index]
        if (header) {
          options.headers[index] = header
        }
      }
    }

    // Legacy: this should be specified as `eventSourceInitDict.https.rejectUnauthorized`,
    // but for now exists as a backwards-compatibility layer
    options.rejectUnauthorized = !(eventSourceInitDict && !eventSourceInitDict.rejectUnauthorized)

    if (eventSourceInitDict && eventSourceInitDict.createConnection !== undefined) {
      options.createConnection = eventSourceInitDict.createConnection
    }

    // If specify http proxy, make the request to sent to the proxy server,
    // and include the original url in path and Host headers
    const useProxy = eventSourceInitDict && eventSourceInitDict.proxy
    if (useProxy) {
      const proxy = parse(eventSourceInitDict.proxy)
      isSecure = proxy.protocol === 'https:'

      options.protocol = isSecure ? 'https:' : 'http:'
      options.path = url
      options.headers.Host = options.host
      options.hostname = proxy.hostname
      options.host = proxy.host
      options.port = proxy.port
    }

    // If https options are specified, merge them into the request options
    if (eventSourceInitDict && eventSourceInitDict.https) {
      for (const optName in eventSourceInitDict.https) {
        if (!httpsOptions.has(optName)) {
          continue
        }

        const option = eventSourceInitDict.https[optName]
        if (option !== undefined) {
          options[optName] = option
        }
      }
    }

    // Pass this on to the XHR
    if (eventSourceInitDict && eventSourceInitDict.withCredentials !== undefined) {
      options.withCredentials = eventSourceInitDict.withCredentials
    }

    if (proxyAgent) {
      options.agent = proxyAgent(url)
    }

    request = (isSecure ? https : http).request(options, res => {
      self.connectionInProgress = false
      // Handle HTTP errors
      if (res.statusCode === 500 || res.statusCode === 502 || res.statusCode === 503 || res.statusCode === 504) {
        _emit('error', new Event('error', { status: res.statusCode, message: res.statusMessage }))
        onConnectionClosed()
        return
      }

      // Handle HTTP redirects
      if (res.statusCode === 301 || res.statusCode === 302 || res.statusCode === 307) {
        if (!res.headers.location) {
          // Server sent redirect response without Location header.
          _emit('error', new Event('error', { status: res.statusCode, message: res.statusMessage }))
          return
        }
        if (res.statusCode === 307) {
          reconnectUrl = url
        }
        url = res.headers.location
        process.nextTick(connect)
        return
      }

      // Debt: make an exception for the invalid status code of 0.
      // All VS Code requests intercepted by Polly.js result in 0 status codes.
      if (res.statusCode !== 200 && res.statusCode !== 0) {
        _emit('error', new Event('error', { status: res.statusCode, message: res.statusMessage }))
        return self.close()
      }

      readyState = EventSource.OPEN
      res.on('close', () => {
        res.removeAllListeners('close')
        res.removeAllListeners('end')
        onConnectionClosed()
      })

      res.on('end', () => {
        res.removeAllListeners('close')
        res.removeAllListeners('end')
        onConnectionClosed()
      })
      _emit('open', new Event('open'))

      // text/event-stream parser adapted from webkit's
      // Source/WebCore/page/EventSource.cpp
      let isFirst = true
      let buf
      let startingPosition = 0
      let startingFieldLength = -1
      res.on('data', chunk => {
        buf = buf ? Buffer.concat([buf, chunk]) : chunk
        if (isFirst && hasBom(buf)) {
          buf = buf.slice(bom.length)
        }

        isFirst = false
        let position = 0
        const length = buf.length

        while (position < length) {
          if (discardTrailingNewline) {
            if (buf[position] === lineFeed) {
              ++position
            }
            discardTrailingNewline = false
          }

          let lineLength = -1
          let fieldLength = startingFieldLength
          var c

          for (let index = startingPosition; lineLength < 0 && index < length; ++index) {
            c = buf[index]
            if (c === colon) {
              if (fieldLength < 0) {
                fieldLength = index - position
              }
            } else if (c === carriageReturn) {
              discardTrailingNewline = true
              lineLength = index - position
            } else if (c === lineFeed) {
              lineLength = index - position
            }
          }

          if (lineLength < 0) {
            startingPosition = length - position
            startingFieldLength = fieldLength
            break
          } else {
            startingPosition = 0
            startingFieldLength = -1
          }

          parseEventStreamLine(buf, position, fieldLength, lineLength)

          position += lineLength + 1
        }

        if (position === length) {
          buf = void 0
        } else if (position > 0) {
          buf = buf.slice(position)
        }
      })
    })

    request.on('error', error => {
      self.connectionInProgress = false
      onConnectionClosed(error.message)
    })

    if (request.setNoDelay) {
      request.setNoDelay(true)
    }
    request.end()
  }

  connect()

  function _emit() {
    if (self.listeners(arguments[0]).length > 0) {
      self.emit.apply(self, arguments)
    }
  }

  this._close = function () {
    if (readyState === EventSource.CLOSED) {
      return
    }
    readyState = EventSource.CLOSED
    if (request.abort) {
      request.abort()
    }
    if (request.xhr && request.xhr.abort) {
      request.xhr.abort()
    }
  }

  function parseEventStreamLine(buf, position, fieldLength, lineLength) {
    if (lineLength === 0) {
      if (data.length > 0) {
        const type = eventName || 'message'
        _emit(
          type,
          new MessageEvent(type, {
            data: data.slice(0, -1), // remove trailing newline
            lastEventId,
            origin: original(url),
          })
        )
        data = ''
      }
      eventName = void 0
    } else if (fieldLength > 0) {
      const noValue = fieldLength < 0
      let step = 0
      const field = buf.slice(position, position + (noValue ? lineLength : fieldLength)).toString()

      if (noValue) {
        step = lineLength
      } else if (buf[position + fieldLength + 1] !== space) {
        step = fieldLength + 1
      } else {
        step = fieldLength + 2
      }
      position += step

      const valueLength = lineLength - step
      const value = buf.slice(position, position + valueLength).toString()

      if (field === 'data') {
        data += value + '\n'
      } else if (field === 'event') {
        eventName = value
      } else if (field === 'id') {
        lastEventId = value
      } else if (field === 'retry') {
        const retry = parseInt(value, 10)
        if (!Number.isNaN(retry)) {
          self.reconnectInterval = retry
        }
      }
    }
  }
}

// module.exports = EventSource

util.inherits(EventSource, events.EventEmitter)
EventSource.prototype.constructor = EventSource // make stacktraces readable
;['open', 'error', 'message'].forEach(method => {
  Object.defineProperty(EventSource.prototype, 'on' + method, {
    /**
     * Returns the current listener
     *
     * @returns {Mixed} the set function or undefined
     * @api private
     */
    get: function get() {
      const listener = this.listeners(method)[0]
      return listener ? (listener._listener ? listener._listener : listener) : undefined
    },

    /**
     * Start listening for events
     *
     * @param {Function} listener the listener
     * @returns {Mixed} the set function or undefined
     * @api private
     */
    set: function set(listener) {
      this.removeAllListeners(method)
      this.addEventListener(method, listener)
    },
  })
})

/**
 * Ready states
 */
Object.defineProperty(EventSource, 'CONNECTING', { enumerable: true, value: 0 })
Object.defineProperty(EventSource, 'OPEN', { enumerable: true, value: 1 })
Object.defineProperty(EventSource, 'CLOSED', { enumerable: true, value: 2 })

EventSource.prototype.CONNECTING = 0
EventSource.prototype.OPEN = 1
EventSource.prototype.CLOSED = 2

/**
 * Closes the connection, if one is made, and sets the readyState attribute to 2 (closed)
 *
 * @see https://developer.mozilla.org/en-US/docs/Web/API/EventSource/close
 * @api public
 */
EventSource.prototype.close = function () {
  this._close()
}

/**
 * Emulates the W3C Browser based WebSocket interface using addEventListener.
 *
 * @param {string} type A string representing the event type to listen out for
 * @param {Function} listener callback
 * @see https://developer.mozilla.org/en/DOM/element.addEventListener
 * @see http://dev.w3.org/html5/websockets/#the-websocket-interface
 * @api public
 */
EventSource.prototype.addEventListener = function addEventListener(type, listener) {
  if (typeof listener === 'function') {
    // store a reference so we can return the original function again
    listener._listener = listener
    this.on(type, listener)
  }
}

/**
 * Emulates the W3C Browser based WebSocket interface using dispatchEvent.
 *
 * @param {Event} event An event to be dispatched
 * @see https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/dispatchEvent
 * @api public
 */
EventSource.prototype.dispatchEvent = function dispatchEvent(event) {
  if (!event.type) {
    throw new Error('UNSPECIFIED_EVENT_TYPE_ERR')
  }
  // if event is instance of an CustomEvent (or has 'details' property),
  // send the detail object as the payload for the event
  this.emit(event.type, event.detail)
}

/**
 * Emulates the W3C Browser based WebSocket interface using removeEventListener.
 *
 * @param {string} type A string representing the event type to remove
 * @param {Function} listener callback
 * @see https://developer.mozilla.org/en/DOM/element.removeEventListener
 * @see http://dev.w3.org/html5/websockets/#the-websocket-interface
 * @api public
 */
EventSource.prototype.removeEventListener = function removeEventListener(type, listener) {
  if (typeof listener === 'function') {
    listener._listener = undefined
    this.removeListener(type, listener)
  }
}

/**
 * W3C Event
 *
 * @see http://www.w3.org/TR/DOM-Level-3-Events/#interface-Event
 * @api private
 */
function EventPolyfill(type, optionalProperties) {
  Object.defineProperty(this, 'type', { writable: false, value: type, enumerable: true })
  if (optionalProperties) {
    for (const f in optionalProperties) {
      if (optionalProperties.hasOwnProperty(f)) {
        Object.defineProperty(this, f, { writable: false, value: optionalProperties[f], enumerable: true })
      }
    }
  }
}

/**
 * W3C MessageEvent
 *
 * @see http://www.w3.org/TR/webmessaging/#event-definitions
 * @api private
 */
function MessageEventPolyfill(type, eventInitDict) {
  Object.defineProperty(this, 'type', { writable: false, value: type, enumerable: true })
  for (const f in eventInitDict) {
    if (eventInitDict.hasOwnProperty(f)) {
      Object.defineProperty(this, f, { writable: false, value: eventInitDict[f], enumerable: true })
    }
  }
}
