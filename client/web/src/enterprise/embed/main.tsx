import '@sourcegraph/shared/src/polyfills'

import { createRoot } from 'react-dom/client'

import { EmbeddedWebApp } from './EmbeddedWebApp'

// It's important to have a root component in a separate file to create a react-refresh boundary and avoid page reload.
// https://github.com/pmmmwh/react-refresh-webpack-plugin/blob/main/docs/TROUBLESHOOTING.md#edits-always-lead-to-full-reload
window.addEventListener('DOMContentLoaded', () => {
    const root = createRoot(document.querySelector('#root')!)

    root.render(<EmbeddedWebApp />)
})
