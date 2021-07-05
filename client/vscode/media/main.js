//@ts-check

// This script will be run within the webview itself
// It cannot access the main VS Code APIs directly.
;(function () {
  const vscode = acquireVsCodeApi()

  // Handle messages sent from the extension to the webview
  window.addEventListener('message', event => {
    console.log('WINDOW message', event)
    const message = event.data // The json data that the extension sent
    switch (message.type) {
      case 'cursor':
        break
      case 'usageClick':
        vscode.postMessage({ type: 'usageSelected', value: 'xyz' })
        break
    }
  })

  const iframe = document.createElement('iframe')
  iframe.src =
    'https://sourcegraph.test:3443/github.com/hashicorp/go-multierror@2004d9dba6b07a5b8d133209244f376680f9d472/-/usage/symbol/gomod/github.com/hashicorp/go-multierror:Append'
  iframe.height = '100%'
  iframe.width = '100%'
  // iframe.sandbox.add('allow-scripts', 'allow-same-origin')
  document.body.appendChild(iframe)
})()
