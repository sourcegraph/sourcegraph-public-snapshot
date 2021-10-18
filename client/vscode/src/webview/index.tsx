import React from 'react'
import { render } from 'react-dom'

import { SearchPage } from './search'

// TODO tsconfig for this dir, vscode typedef (for postmessage)
// Wrap with comlink, pass comlink.Remote<SGVSCodeExtensionAPI> to <SearchWebviewApp/>
const Main: React.FC = () => {
    const route: 'search' | 'settings' | null = 'search'

    if (route === null) {
        return (
            <div>
                <h1>Webview under construction</h1>
            </div>
        )
    }

    return <SearchPage />
}
render(<Main />, document.querySelector('#root'))
