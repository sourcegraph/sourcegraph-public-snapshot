import { render } from 'react-dom'

import { setLinkComponent, AnchorLink } from '@sourcegraph/wildcard'

import { App } from './App'

setLinkComponent(AnchorLink)

const node = document.querySelector('#main') as HTMLDivElement
render(<App />, node)
