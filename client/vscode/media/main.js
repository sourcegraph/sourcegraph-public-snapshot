//@ts-check

// This script will be run within the webview itself
// It cannot access the main VS Code APIs directly.
;(function () {
  const vscode = acquireVsCodeApi()

  const EMPTY_IFRAME_SRC =
    'https://sourcegraph.test:3443/github.com/hashicorp/go-multierror@2004d9dba6b07a5b8d133209244f376680f9d472/-/usage'

  const iframe = document.createElement('iframe')
  iframe.src = EMPTY_IFRAME_SRC
  iframe.height = '100%'
  iframe.width = '100%'
  // iframe.sandbox.add('allow-scripts', 'allow-same-origin')
  document.body.appendChild(iframe)

  // Handle messages sent from the extension to the webview
  window.addEventListener('message', event => {
    console.log('WINDOW message', event)
    const message = event.data // The json data that the extension sent
    switch (message.type) {
      case 'cursor':
        if (message.url) {
          iframe.src = message.url
        } else {
          iframe.src = EMPTY_IFRAME_SRC
        }
        break
      case 'usageClick':
        vscode.postMessage({ type: 'usageSelected', value: 'xyz' })
        break
    }
  })
})()
