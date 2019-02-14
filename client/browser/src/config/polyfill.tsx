// This gets automatically expanded into imports that only pick what we need.
import '@babel/polyfill'

// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
// The polyfill does not expose createObjectURL, which we need for creating data: URIs for Web
// Workers. So retain it.
//
// tslint:disable-next-line:no-unbound-method
const createObjectURL = window.URL ? window.URL.createObjectURL : null

import { URL, URLSearchParams } from 'whatwg-url'

const GLOBAL = global as any
GLOBAL.URL = URL
GLOBAL.URLSearchParams = URLSearchParams
if (!window.URL.createObjectURL && createObjectURL) {
    window.URL.createObjectURL = createObjectURL
}
