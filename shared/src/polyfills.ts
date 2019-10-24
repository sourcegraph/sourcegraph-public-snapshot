// Commonly needed by extensions (used by vscode-jsonrpc)
import 'core-js/web/immediate'

import 'symbol-observable'

// This gets automatically expanded into imports that only pick what we need.
import 'core-js/stable'
import 'regenerator-runtime/runtime'

// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import { URL, URLSearchParams } from 'whatwg-url'
Object.assign(URL, self.URL) // keep static methods like URL.createObjectURL()
Object.assign(self, { URL, URLSearchParams })
