import './sandbox.scss'

import { createRoot } from 'react-dom/client'

import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { UsageExamplesBox } from '../src/usageExamples/UsageExamplesBox'

setLinkComponent(AnchorLink)

const container = document.querySelector('#react-container')
if (container) {
    const root = createRoot(container)
    root.render(<UsageExamplesBox />)
}
