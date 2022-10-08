import './index.scss'

import { createRoot } from 'react-dom/client'
import { UsageExamplesBox } from 'usageExamples/UsageExamplesBox'

import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

setLinkComponent(AnchorLink)

const container = document.querySelector('#react-container')
if (container) {
    const root = createRoot(container)
    root.render(<UsageExamplesBox />)
}
