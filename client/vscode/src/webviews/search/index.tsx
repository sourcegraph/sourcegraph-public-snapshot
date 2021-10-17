import React from 'react'
import { render } from 'react-dom'

import { areExtensionsSame } from '@sourcegraph/shared/src/extensions/extensions'

// TODO tsconfig for this dir, vscode typedef (for postmessage)
// Wrap with comlink, pass comlink.Remote<SGVSCodeExtensionAPI> to <SearchWebviewApp/>
setInterval(() => {
    console.log('the search webview js file!')
    const shouldDownloadExtensions = !areExtensionsSame([{ id: 'old' }], [{ id: 'new' }])
    console.log({ shouldDownloadExtensions })
}, 2000)

// window.addEventListener('message', event => {
//     console.log('msg', event.data)
// })

render(
    <div>
        <h1>Test h1</h1>
        <input type="text" placeholder="NEW SEARCH" />
    </div>,
    document.querySelector('#root')
)
