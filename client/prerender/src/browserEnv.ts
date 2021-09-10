const EventListener = { addEventListener: () => null, removeEventListener: () => null }
global.window = global
global.navigator = { platform: '', userAgent: '' }
global.location = {}
global.dispatchEvent = () => {}
global.context = require('./jscontext').JSCONTEXT
global.matchMedia = () => ({
    // TODO(sqs): hack, this is only used by the theme code so this means the default is dark theme
    matches: true,
    ...EventListener,
})
global.addEventListener = EventListener.addEventListener
global.removeEventListener = EventListener.removeEventListener

global.localStorage = {
    getItem: () => null,
    setItem: () => {},
    removeItem: () => {},
}
const ELEMENT = {
    setAttribute: () => {},
    classList: {
        add: () => {},
    },
    style: {},
    append: () => {},
}
global.document = {
    querySelector: () => null,
    createEvent: () => ({ initCustomEvent: () => {}, ...EventListener }),
    documentElement: ELEMENT,
    createElement: () => ELEMENT,
    head: ELEMENT,
    ...EventListener,
}
global.Node = {}
global.navigator = window.navigator
global.location = window.location
global.Element = {
    scroll: null,
    prototype: {
        matches: () => false,
        scroll: () => {},
    },
}
window.Element = global.Element
global.self = { ...window, close: () => {} }
global.fetch = require('cross-fetch')

import crypto from 'crypto'

Object.defineProperty(global.self, 'crypto', {
    value: {
        subtle: crypto.webcrypto.subtle,
    },
})
