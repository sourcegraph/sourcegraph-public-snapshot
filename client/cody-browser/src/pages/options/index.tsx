import Options from '@pages/options/Options'
import { createRoot } from 'react-dom/client'

import '@pages/options/index.css'
import '@assets/styles/tailwind.css'

function init() {
    const rootContainer = document.querySelector('#__root')
    if (!rootContainer) throw new Error("Can't find Options root element")
    const root = createRoot(rootContainer)
    root.render(<Options />)
}

init()
