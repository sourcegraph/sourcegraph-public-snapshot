// This gets automatically expanded into
// imports that only pick what we need
import '@babel/polyfill'

// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import { URL, URLSearchParams } from 'whatwg-url'
// The polyfill does not expose createObjectURL, which we need for creating data: URIs for Web
// Workers. So retain it.
URL.createObjectURL = window.URL.createObjectURL
Object.assign(window, { URL, URLSearchParams })
