import 'symbol-observable'

// This gets automatically expanded into
// imports that only pick what we need
import '@babel/polyfill'

// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import { URL, URLSearchParams } from 'whatwg-url'
if (self.URL && self.URL.createObjectURL && self.URL.revokeObjectURL) {
    ;(URL as Window['URL']).createObjectURL = self.URL.createObjectURL
    ;(URL as Window['URL']).revokeObjectURL = self.URL.revokeObjectURL
}
Object.assign(self, { URL, URLSearchParams })
