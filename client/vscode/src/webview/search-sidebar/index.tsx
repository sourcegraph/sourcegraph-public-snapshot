import React from 'react'
import { render } from 'react-dom'

import { AnchorLink, setLinkComponent } from '@sourcegraph/shared/src/components/Link'

// TODO: load extension host

setLinkComponent(AnchorLink)

const Main: React.FC = () => {
    console.log('rendering webview')
    return (
        <div>
            <h1>WIP</h1>
        </div>
    )
}

render(<Main />, document.querySelector('#root'))
