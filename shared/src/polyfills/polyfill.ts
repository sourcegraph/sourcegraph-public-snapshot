// This gets expanded into only the imports we need by @babel/preset-env
import 'core-js/stable'

// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import 'core-js/web/url'

// Commonly needed by extensions (used by vscode-jsonrpc)
import 'core-js/web/immediate'

// Avoids issues with RxJS interop
import 'symbol-observable'
